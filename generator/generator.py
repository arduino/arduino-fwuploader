#!/usr/bin/env python3

# Copyright 2021 Arduino SA
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published
# by the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

import argparse
import json
import hashlib
import shutil
from pathlib import Path

DOWNLOAD_URL = "https://downloads.arduino.cc/arduino-fwuploader"


# handle firmware name
def get_firmware_file(module, simple_fqbn, version):
    firmware_full_path = Path(__file__).parent.parent / "firmwares" / module / version
    fqbn_specific_file_name = f"{module}-{simple_fqbn}.bin"
    if (firmware_file := firmware_full_path / fqbn_specific_file_name).exists():
        return firmware_file
    return firmware_full_path / f"{module}.bin"


# Generates file SHA256
def sha2(file_path):
    with open(file_path, "rb") as f:
        return hashlib.sha256(f.read()).hexdigest()


def split_property_and_drop_first_level(s):
    (k, v) = s.strip().split("=", maxsplit=1)
    k = ".".join(k.split(".", maxsplit=1)[1:])
    return (k, v)


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


def generate_new_boards_json(input_data):
    # init the boards dict
    boards = {}
    for fqbn, data in input_data.items():
        simple_fqbn = fqbn.replace(":", ".")

        # populate the boards dict
        boards[fqbn] = {}
        boards[fqbn]["fqbn"] = fqbn
        module = data["moduleName"]
        boards[fqbn]["firmware"] = []
        for firmware_version in data["versions"]:
            firmware_file = get_firmware_file(module, simple_fqbn, firmware_version)
            boards[fqbn]["firmware"].append(create_firmware_data(firmware_file, module, firmware_version))
        boards[fqbn]["uploader_plugin"] = data["uploader_plugin"]
        boards[fqbn]["additional_tools"] = data["additional_tools"]
        boards[fqbn]["module"] = module
        boards[fqbn]["name"] = data["name"]

    boards_json = []
    for _, b in boards.items():
        boards_json.append(b)

    return boards_json


if __name__ == "__main__":
    parser = argparse.ArgumentParser(prog="generator.py")

    with open("new_boards.json", "r") as f:
        boards = json.load(f)

    boards_json = generate_new_boards_json(boards)

    Path("boards").mkdir(exist_ok=True)

    with open("boards/plugin_firmware_index.json", "w") as f:
        json.dump(boards_json, f, indent=2)
