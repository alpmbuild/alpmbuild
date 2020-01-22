package lib

/*
   alpmbuild — a tool to build arch packages from RPM specfiles

   Copyright (C) 2020  Carson Black

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

type packageContext struct {
	Name    string `macro:"name" key:"name:"`
	Version string `macro:"version" key:"version:"`
	Release string `macro:"release" key:"release:"`
	Summary string `macro:"summary" key:"summary:"`
	License string `macro:"license" key:"license:"`
	URL     string `macro:"url" key:"url:"`

	Sources []string
}
