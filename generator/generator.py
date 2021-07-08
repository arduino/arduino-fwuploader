#!/usr/bin/env python3

import argparse
import subprocess
import sys
import json
import hashlib
import shutil
import os
from pathlib import Path

DOWNLOAD_URL = "https://downloads.arduino.cc/arduino-fwuploader"

# Runs arduino-cli, doesn't handle errors at all because am lazy
def arduino_cli(cli_path, args=[]):
    res = subprocess.run([cli_path, *args], capture_output=True, text=True)
    return res.stdout


# Generates file SHA256
def sha2(file_path):
    with open(file_path, "rb") as f:
        return hashlib.sha256(f.read()).hexdigest()


def split_property_and_drop_first_level(s):
    (k, v) = s.strip().split("=", maxsplit=1)
    k = ".".join(k.split(".", maxsplit=1)[1:])
    return (k, v)


# Generate and copy loader Sketch binary data for specified board
def create_loader_data(simple_fqbn, binary):
    loader_path = f"firmwares/loader/{simple_fqbn}/loader{binary.suffix}"
    loader = Path(__file__).parent / loader_path
    loader.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(binary, loader)

    file_hash = sha2(loader)

    return {
        "url": f"{DOWNLOAD_URL}/{loader_path}",
        "checksum": f"SHA-256:{file_hash}",
        "size": f"{loader.stat().st_size}",
    }


# Generate and copy all firmware binary data for specified board
def create_firmware_data(binary, module, version):
    binary_name = binary.name
    firmware_path = f"firmwares/{module}/{version}/{binary_name}"
    firmware = Path(__file__).parent / firmware_path
    firmware.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(binary, firmware)

    file_hash = sha2(firmware)

    return {
        "version": version,
        "url": f"{DOWNLOAD_URL}/{firmware_path}",
        "checksum": f"SHA-256:{file_hash}",
        "size": f"{firmware.stat().st_size}",
        "module": module,
    }


def get_uploader_id(tools, tool_name):
    for t in tools:
        if t["name"] == tool_name:
            packager = t["packager"]
            name = t["name"]
            version = t["version"]
            return f"{packager}:{name}@{version}"


def create_upload_data(fqbn, installed_cores):  # noqa: C901
    upload_data = {}
    # Assume we're on Linux
    arduino15 = Path.home() / ".arduino15"

    board_id = fqbn.split(":")[2]
    core_id = ":".join(fqbn.split(":")[:2])

    # Get the core install dir
    core = installed_cores[core_id]
    (maintainer, arch) = core_id.split(":")
    core_install_dir = arduino15 / "packages" / maintainer / "hardware" / arch / core["installed"]

    with open(core_install_dir / "boards.txt") as f:
        boards_txt = f.readlines()

    board_upload_data = {}
    for line in boards_txt:
        if line.startswith(f"{board_id}."):
            (k, v) = split_property_and_drop_first_level(line)
            board_upload_data[k] = v

    tool = board_upload_data["upload.tool"]

    with open(core_install_dir / "platform.txt") as f:
        platform_txt = f.readlines()

    platform_upload_data = {}
    for line in platform_txt:
        if line.startswith(f"tools.{tool}."):
            (k, v) = split_property_and_drop_first_level(line)
            platform_upload_data[k] = v

    # We assume the installed.json exist
    with open(core_install_dir / "installed.json") as f:
        installed_json_data = json.load(f)

    if f"{tool}.cmd" in platform_upload_data:
        tool_executable_generic = platform_upload_data[f"{tool}.cmd"]
        tool_executable_linux = platform_upload_data.get(f"{tool}.cmd.linux", tool_executable_generic)
        tool_executable_windows = platform_upload_data.get(f"{tool}.cmd.windows", "")
        tool_executable_macosx = platform_upload_data.get(f"{tool}.cmd.macosx", "")
        tool_name = tool_executable_generic
    elif f"{tool}.cmd.path" in platform_upload_data:
        tool_executable_generic = "/".join(platform_upload_data[f"{tool}.cmd.path"].split("/")[1:])
        tool_executable_linux = platform_upload_data.get(f"{tool}.cmd.path.linux", tool_executable_generic)
        tool_executable_windows = platform_upload_data.get(f"{tool}.cmd.path.windows", "")
        tool_executable_macosx = platform_upload_data.get(f"{tool}.cmd.path.macosx", "")
        tool_name = tool_executable_generic.split("/")[-1]

    tool_config_path = ""
    if f"{tool}.config.path" in platform_upload_data:
        tool_config_path = "/".join(platform_upload_data[f"{tool}.config.path"].split("/")[1:])

    if tool_name == "rp2040load":
        tool_name = "rp2040tools"

    tools = installed_json_data["packages"][0]["platforms"][0]["toolsDependencies"]
    upload_data["uploader"] = get_uploader_id(tools, tool_name)

    if "upload.use_1200bps_touch" in board_upload_data:
        upload_data["upload.use_1200bps_touch"] = bool(board_upload_data["upload.use_1200bps_touch"])

    if "upload.wait_for_upload_port" in board_upload_data:
        upload_data["upload.wait_for_upload_port"] = bool(board_upload_data["upload.wait_for_upload_port"])

    # Get the command used to upload and modifies it a bit
    command = (
        platform_upload_data[f"{tool}.upload.pattern"]
        .replace("{path}/{cmd}", "{uploader}")
        .replace("{cmd.path}", "{uploader}")
        .replace("{build.path}/{build.project_name}", "{loader.sketch}")
        .replace("{config.path}", f"{{tool_dir}}/{tool_config_path}")
        .replace('\\"', "")
    )

    if fqbn == "arduino:megaavr:uno2018":
        # Long story short if we don't do this we'd have to host also the bootloader
        # for the Uno WiFi rev2 and we don't want to, so we just remove this field
        # and use a precompiled Loader Sketh binary that includes the bootloader.
        command = command.replace("{upload.extra_files}", "")

    # Get the rest of the params
    params = {}
    for k, v in platform_upload_data.items():
        if f"{tool}.upload.params." in k:
            param = k.split(".")[-1]
            params[f"upload.{param}"] = v
        elif f"{tool}.upload." in k:
            k = ".".join(k.split(".")[1:])
            params[k] = v

    # Prepare the command
    for k, v in {**board_upload_data, **params}.items():
        command = command.replace(f"{{{k}}}", v)

    # This is ugly as hell and I don't care
    upload_data["uploader.command"] = {}
    if tool_executable_linux:
        upload_data["uploader.command"]["linux"] = command.replace(
            "{uploader}", f"{{tool_dir}}/{tool_executable_linux}"
        )

    if tool_executable_windows:
        upload_data["uploader.command"]["windows"] = command.replace(
            "{uploader}", f"{{tool_dir}}\\{tool_executable_windows}"
        )

    if tool_executable_macosx:
        upload_data["uploader.command"]["macosx"] = command.replace(
            "{uploader}", f"{{tool_dir}}/{tool_executable_macosx}"
        )

    return upload_data


def generate_boards_json(input_data, arduino_cli_path):
    boards = {
        "arduino:samd:mkr1000": {"fqbn": "arduino:samd:mkr1000", "firmware": []},
        "arduino:samd:mkrwifi1010": {
            "fqbn": "arduino:samd:mkrwifi1010",
            "firmware": [],
        },
        "arduino:samd:nano_33_iot": {
            "fqbn": "arduino:samd:nano_33_iot",
            "firmware": [],
        },
        "arduino:samd:mkrvidor4000": {
            "fqbn": "arduino:samd:mkrvidor4000",
            "firmware": [],
        },
        "arduino:megaavr:uno2018": {"fqbn": "arduino:megaavr:uno2018", "firmware": []},
        "arduino:mbed_nano:nanorp2040connect": {
            "fqbn": "arduino:mbed_nano:nanorp2040connect",
            "firmware": [],
        },
    }

    # Gets the installed cores
    res = arduino_cli(cli_path=arduino_cli_path, args=["core", "list", "--format", "json"])
    installed_cores = {c["id"]: c for c in json.loads(res)}

    # Verify all necessary cores are installed
    # TODO: Should we check that the latest version is installed too?
    for fqbn in boards.keys():
        core_id = ":".join(fqbn.split(":")[:2])
        if core_id not in installed_cores:
            print(f"Board {fqbn} is not installed, install its core {core_id}")
            sys.exit(1)

    for fqbn, data in input_data.items():
        simple_fqbn = fqbn.replace(":", ".")

        binary_path = f"../firmwares/loader/{simple_fqbn}/"
        binary = (
            Path(__file__).parent / binary_path / os.listdir(binary_path)[0]
        )  # there's only one loader bin in every fqbn dir
        boards[fqbn]["loader_sketch"] = create_loader_data(simple_fqbn, binary)

        for firmware_version in data["versions"]:

            # handle firmware name
            if fqbn == "arduino:megaavr:uno2018":
                firmware = "NINA_W102-arduino.megaavr.uno2018.bin"
            elif fqbn == "arduino:mbed_nano:nanorp2040connect":
                firmware = "NINA_W102-arduino.mbed_nano.nanorp2040connect.bin"
            elif fqbn == "arduino:samd:mkr1000":
                firmware = "m2m_aio_3a0-arduino.samd.mkr1000.bin"
            else:
                firmware = "NINA_W102.bin"
            module = data["moduleName"]
            firmware_path = f"firmwares/{module}/{firmware_version}/{firmware}"
            binary = Path(__file__).parent / ".." / firmware_path
            boards[fqbn]["firmware"].append(create_firmware_data(binary, module, firmware_version))
            boards[fqbn]["module"] = module

        res = arduino_cli(
            cli_path=arduino_cli_path,
            args=["board", "search", fqbn, "--format", "json"],
        )
        # Gets the board name
        for board in json.loads(res):
            if board["fqbn"] == fqbn:
                boards[fqbn]["name"] = board["name"]
                break

        boards[fqbn].update(create_upload_data(fqbn, installed_cores))

    boards_json = []
    for _, b in boards.items():
        boards_json.append(b)

    return boards_json


if __name__ == "__main__":
    parser = argparse.ArgumentParser(prog="generator.py")
    parser.add_argument(
        "-a",
        "--arduino-cli",
        default="arduino-cli",
        help="Path to arduino-cli executable",
        required=True,
    )
    args = parser.parse_args(sys.argv[1:])

    # raw_boards.json has been generated using --get_available_for FirmwareUploader (version 0.1.8) flag.
    # It has been edited a bit to better handle parsing.
    with open("boards.json", "r") as f:
        boards = json.load(f)

    boards_json = generate_boards_json(boards, args.arduino_cli)

    Path("boards").mkdir()

    with open("boards/module_firmware_index.json", "w") as f:
        json.dump(boards_json, f, indent=2)

# board_index.json must be formatted like so:
#
# {
#     "name": "MKR 1000",
#     "fqbn": "arduino:samd:mkr1000",
#     "module": "WINC_1500",
#     "firmware": [
#         {
#             "version": "19.6.1",
#             "url": "https://downloads.arduino.cc/firmwares/WINC_1500/19.6.1/m2m_aio_3a0.bin",
#             "checksum": "SHA-256:de0c6b1621aa15996432559efb5d8a29885f62bde145937eee99883bfa129f97",
#             "size": "359356",
#         },
#         {
#             "version": "19.5.4",
#             "url": "https://downloads.arduino.cc/firmwares/WINC_1500/19.5.4/m2m_aio_3a0.bin",
#             "checksum": "SHA-256:71e5a805e60f96e6968414670d8a414a03cb610fd4b020f47ab53f5e1ff82a13",
#             "size": "413604",
#         },
#     ],
#     "loader_sketch": {
#         "url": "https://downloads.arduino.cc/firmwares/loader/arduino.samd.mkr1000/loader.bin",
#         "checksum": "SHA-256:71e5a805e60f96e6968414670d8a414a03cb610fd4b020f47ab53f5e1ff82a13",
#         "size": "39287",
#     },
#     "uploader": "arduino:bossac@1.7.0",
#     "uploader.command": "{uploader} --port={upload.port} -U true -i -e -w -v {loader.sketch} -R",
#     "uploader.requires_1200_bps_touch": "true",
#     "uploader.requires_port_change": "true",
# }
