import json
import hashlib
import shutil
from pathlib import Path

DOWNLOAD_URL = "https://downloads.arduino.cc"
FQBNS = {
    "mkr1000": "arduino:samd:mkr1000",
    "mkrwifi1010": "arduino:samd:mkrwifi1010",
    "nano_33_iot": "arduino:samd:nano_33_iot",
    "mkrvidor4000": "arduino:samd:mkrvidor4000",
    "uno2018": "arduino:megaavr:uno2018",
    "mkrnb1500": "arduino:samd:mkrnb1500",
    "nanorp2040connect": "arduino:mbed_nano:nanorp2040connect",
}


# Generates file SHA256
def sha2(file_path):
    with open(file_path, "rb") as f:
        return hashlib.sha256(f.read()).hexdigest()


# Generate and copy loader Sketch binary data for specified board
def create_loader_data(simple_fqbn, binary):
    loader_path = f"firmwares/loader/{simple_fqbn}/loader.bin"
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
def create_firmware_data(simple_fqbn, binary, module, version):
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


def generate_boards_json(input_data):
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

    for pseudo_fqbn, data in input_data.items():
        fqbn = FQBNS[pseudo_fqbn]
        simple_fqbn = fqbn.replace(":", ".")

        for _, v in data.items():
            item = v[0]
            binary = Path(item["Path"])

            if item["IsLoader"]:
                boards[fqbn]["loader_sketch"] = create_loader_data(simple_fqbn, binary)
            else:
                module, version = item["version"].split("/")
                boards[fqbn]["firmware"].append(
                    create_firmware_data(simple_fqbn, binary, module, version)
                )

    # TODO: Run arduino-cli to get board names and other things?

    boards_json = []
    for _, b in boards.items():
        boards_json.append(b)


if __name__ == "__main__":
    # raw_boards.json has been generated using --get_available_for FirmwareUploader flag.
    # It has been edited a bit to better handle parsing.
    with open("raw_boards.json", "r") as f:
        raw_boards = json.load(f)

    boards_json = generate_boards_json(raw_boards)

    with open("boards.json", "w") as f:
        json.dump(boards_json, f, indent=2)

# boards.json must be formatted like so:
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
