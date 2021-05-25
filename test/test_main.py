#   FirmwareUploader
#   Copyright (c) 2021 Arduino LLC.  All right reserved.

#   This library is free software; you can redistribute it and/or
#   modify it under the terms of the GNU Lesser General Public
#   License as published by the Free Software Foundation; either
#   version 2.1 of the License, or (at your option) any later version.

#   This library is distributed in the hope that it will be useful,
#   but WITHOUT ANY WARRANTY; without even the implied warranty of
#   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
#   Lesser General Public License for more details.

#   You should have received a copy of the GNU Lesser General Public
#   License along with this library; if not, write to the Free Software
#   Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA

import json
import semver

def test_help(run_command):
    result = run_command("help")
    assert result.ok
    assert result.stderr == ""
    assert "Usage" in result.stdout


def test_version(run_command):
    result = run_command("version")
    assert result.ok
    assert "Version:" in result.stdout
    assert "Commit:" in result.stdout
    assert "Date:" in result.stdout
    assert "" == result.stderr

    result = run_command("version --format json")
    assert result.ok
    parsed_out = json.loads(result.stdout)
    assert parsed_out.get("Application", False) == "FirmwareUploader"
    version = parsed_out.get("VersionString", False)
    assert semver.VersionInfo.isvalid(version=version) or "git-snapshot" in version or "nightly" in version
    assert isinstance(parsed_out.get("Commit", False), str)
    assert isinstance(parsed_out.get("Date", False), str)
