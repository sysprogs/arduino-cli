/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package common

// Index represents a generic index parsed from an index file.
type Index interface {
	CreateStatusContext() (StatusContext, error) // CreateStatusContext creates a status context with this index data.
}

// StatusContext represents a generic status context, created from an Index.
type StatusContext interface {
	Names() []string               // Names Returns an array with all the names of the items.
	Items() map[string]interface{} // Items Returns a map of all items with their names.
}

// Release represents a generic release.
type Release interface {
	ArchivePath() (string, error) // ArchivePath returns the fullPath of the Archive of this release.
	ExpectedChecksum() string     // Checksum returns the expected checksum for this release.
}