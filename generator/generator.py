import argparse
import subprocess
import sys
import json
import hashlib
import shutil
from pathlib import Path

DOWNLOAD_URL = "https://downloads.arduino.cc/arduino-fwuploader"
FQBNS = {
    "mkr1000": "arduino:samd:mkr1000",
    "mkrwifi1010": "arduino:samd:mkrwifi1010",
    "nano_33_iot": "arduino:samd:nano_33_iot",
    "mkrvidor4000": "arduino:samd:mkrvidor4000",
    "uno2018": "arduino:megaavr:uno2018",
    "mkrnb1500": "arduino:samd:mkrnb1500",
    "nanorp2040connect": "arduino:mbed_nano:nanorp2040connect",
}


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
    }


def get_uploader_id(tools, tool_executable):
    for t in tools:
        if t["name"] == tool_executable:
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
        tool_executable = platform_upload_data[f"{tool}.cmd"]
    elif f"{tool}.cmd.path" in platform_upload_data:
        tool_executable = platform_upload_data[f"{tool}.cmd.path"].split("/")[-1]

    if tool_executable == "rp2040load":
        tool_executable = "rp2040tools"

    tools = installed_json_data["packages"][0]["platforms"][0]["toolsDependencies"]
    upload_data["uploader"] = get_uploader_id(tools, tool_executable)

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

    upload_data["uploader.command"] = command

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
        "arduino:samd:mkrnb1500": {"fqbn": "arduino:samd:mkrnb1500", "firmware": []},
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

    for pseudo_fqbn, data in input_data.items():
        fqbn = FQBNS[pseudo_fqbn]
        simple_fqbn = fqbn.replace(":", ".")

        for _, v in data.items():
            item = v[0]
            binary = Path(__file__).parent / ".." / item["Path"]

            if item["IsLoader"]:
                boards[fqbn]["loader_sketch"] = create_loader_data(simple_fqbn, binary)
            else:
                module, version = item["version"].split("/")
                boards[fqbn]["firmware"].append(create_firmware_data(binary, module, version))
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

    # raw_boards.json has been generated using --get_available_for FirmwareUploader flag.
    # It has been edited a bit to better handle parsing.
    with open("raw_boards.json", "r") as f:
        raw_boards = json.load(f)

    boards_json = generate_boards_json(raw_boards, args.arduino_cli)

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
