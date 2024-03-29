#   arduino-fwuploader
#   Copyright (c) 2021 Arduino LLC.  All right reserved.

#   This program is free software: you can redistribute it and/or modify
#   it under the terms of the GNU Affero General Public License as published
#   by the Free Software Foundation, either version 3 of the License, or
#   (at your option) any later version.

#   This program is distributed in the hope that it will be useful,
#   but WITHOUT ANY WARRANTY; without even the implied warranty of
#   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#   GNU Affero General Public License for more details.

#   You should have received a copy of the GNU Affero General Public License
#   along with this program.  If not, see <https://www.gnu.org/licenses/>.

import json
import semver
import dateutil.parser


def test_version(run_command):
    result = run_command(cmd=["version"])
    assert result.ok
    output_list = result.stdout.strip().split(sep=" ")
    assert output_list[0] == "arduino-fwuploader"
    assert output_list[1] == "Version:"
    version = output_list[2]
    assert semver.VersionInfo.isvalid(version=version) or version == "git-snapshot" or "nightly" in version
    assert output_list[3] == "Commit:"
    assert isinstance(output_list[4], str)
    assert output_list[5] == "Date:"
    assert dateutil.parser.isoparse(output_list[6])
    assert "" == result.stderr

    result = run_command(cmd=["version", "--format", "json"])
    assert result.ok
    parsed_out = json.loads(result.stdout)
    assert parsed_out.get("Application", False) == "arduino-fwuploader"
    version = parsed_out.get("VersionString", False)
    assert semver.VersionInfo.isvalid(version=version) or "git-snapshot" in version or "nightly" in version
    assert parsed_out.get("Commit", False) != ""
    assert isinstance(parsed_out.get("Commit", False), str)
    assert parsed_out.get("Date") != ""
    assert isinstance(parsed_out.get("Date", False), str)
