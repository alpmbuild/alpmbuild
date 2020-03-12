package lib

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

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

// These directives aren't sections, but should cause the stages
// to switch to NoStage
var otherDirectives []string = []string{
	"%package",
}

type Stage int

const (
	NoStage Stage = iota + 1
	PrepareStage
	BuildStage
	InstallStage
	CheckStage
	FileStage

	PreInstallStage
	PostInstallStage
	PreUpgradeStage
	PostUpgradeStage
	PreRemoveStage
	PostRemoveStage

	ChangelogStage

	IfTrueStage
	IfFalseStage
)

var PossibleKeys = []string{
	"Name:",
	"Summary:",
	"License:",
	"URL:",
	"Requires:",
	"BuildRequires:",
	"Recommends:",
	"Version:",
	"Release:",
	"Epoch:",
	"EpochVersionRelease:",
	"EpoVerRel:",
	"EVR:",
	"Provides:",
	"Conflicts:",
	"Replaces:",
	"ExclusiveArch:",
	"Groups:",
	"CheckRequires:",
}

var PossibleDirectives = []string{
	"NoFileCheck",
	"ReasonFor",
}

type CompressionType struct {
	Suffix string
	Flag   string
}

var CompressionTypes = map[string]CompressionType{
	"gz": CompressionType{
		Suffix: "gz",
		Flag:   "-z",
	},
	"xz": CompressionType{
		Suffix: "xz",
		Flag:   "-J",
	},
	"zstd": CompressionType{
		Suffix: "zst",
		Flag:   "--zstd",
	},
	"bz2": CompressionType{
		Suffix: "bz2",
		Flag:   "-j",
	},
}

var packageFields = []string{
	"requires:",
	"recommends:",
	"buildrequires:",
	"provides:",
	"conflicts:",
	"replaces:",
	"checkrequires:",
}

type HashType int

const (
	NoHash HashType = iota
	Sha1Hash
	Sha224Hash
	Sha256Hash
	Sha384Hash
	Sha512Hash
	Md5Hash
)

var hashTypes = []string{
	"sha1",
	"sha224",
	"sha256",
	"sha384",
	"sha512",
	"md5",
	"sig",
	"key",
	"keyserver",
}

type Source struct {
	URL             string
	Rename          string
	Md5             string
	Sha1            string
	Sha256          string
	Sha224          string
	Sha384          string
	Sha512          string
	GPGSignatureURL string
	GPGKeys         []string
	GPGKeyservers   []string
}

type PackageContext struct {
	// Single-value fields with relatively standard behaviour.
	Name    string `macro:"name" key:"name:" pkginfo:"pkgname"`
	Summary string `macro:"summary" key:"summary:" pkginfo:"pkgdesc"`
	License string `macro:"license" key:"license:" pkginfo:"pkglicense"`
	URL     string `macro:"url" key:"url:" pkginfo:"url"`
	Epoch   string `macro:"epoch" key:"epoch:" pkginfo:"epoch"`

	// Array fields with relatively standard behaviour.
	Requires      []string `keyArray:"requires:" pkginfo:"depend"`
	CheckRequires []string `keyArray:"checkrequires:" pkginfo:"checkdepend"`
	Recommends    []string `keyArray:"recommends:" pkginfo:"optdepend"`
	BuildRequires []string `keyArray:"buildrequires:" pkginfo:"makedepend"`
	Provides      []string `keyArray:"provides:" pkginfo:"provides"`
	Conflicts     []string `keyArray:"conflicts:" pkginfo:"conflicts"`
	Replaces      []string `keyArray:"replaces:" pkginfo:"replaces"`
	Groups        []string `keyArray:"groups:" pkginfo:"group"`
	ExclusiveArch []string `keyArray:"exclusivearch:"`

	// Nonstandard single-value fields
	Version string `macro:"version" key:"version:"`
	Release string `macro:"release" key:"release:"`

	// Nonstandard array fields
	Sources   []Source
	Patches   []Source
	Changelog []string
	Files     []string
	Backup    []string `keyArray:"backup:" pkginfo:"backup"`

	// Command fields
	Commands struct {
		Prepare []string
		Build   []string
		Install []string
		Check   []string
	}

	// Scriptlet fields
	Scriptlets struct {
		PreInstall  []string
		PostInstall []string
		PreUpgrade  []string
		PostUpgrade []string
		PreRemove   []string
		PostRemove  []string
	}

	// Other fields
	IsSubpackage  bool
	parentPackage *PackageContext
	Subpackages   map[string]PackageContext
	Reasons       map[string]string
}

func (pkg PackageContext) GetNevra() string {
	uname, _ := exec.Command("uname", "-m").CombinedOutput()
	unameString := strings.TrimSpace(string(uname))
	if pkg.Epoch != "" {
		return fmt.Sprintf("%s-%s:%s-%s-%s", pkg.Name, pkg.Epoch, pkg.Version, pkg.Release, unameString)
	}
	return fmt.Sprintf("%s-%s-%s-%s", pkg.Name, pkg.Version, pkg.Release, unameString)
}

func (pkg PackageContext) GetNevr() string {
	if pkg.Epoch != "" {
		return fmt.Sprintf("%s-%s:%s-%s", pkg.Name, pkg.Epoch, pkg.Version, pkg.Release)
	}
	return fmt.Sprintf("%s-%s-%s", pkg.Name, pkg.Version, pkg.Release)
}

func (pkg PackageContext) GeneratePackageInfo() {
	outputStatus("Generating package info for " + highlight(pkg.GetNevra()) + "...")
	home, err := os.UserHomeDir()
	if err != nil {
		outputError(fmt.Sprintf("Failed to generate pkginfo:\n%s", err.Error()))
	}
	pkgdir := ""
	if !pkg.IsSubpackage {
		pkgdir = filepath.Join(home, "alpmbuild/package")
	} else {
		pkgdir = filepath.Join(home, "alpmbuild/subpackages", pkg.GetNevra())
	}
	os.Chdir(pkgdir)

	packageInfo := "# Generated by alpmbuild"

	fields := reflect.TypeOf(pkg)
	num := fields.NumField()

	for i := 0; i < num; i++ {
		field := fields.Field(i)

		if packageInfoKey := field.Tag.Get("pkginfo"); packageInfoKey != "" {
			// These are the single-value keys, such as Name, Version, Release, and Summary
			if field.Tag.Get("key") != "" {
				// We assert that packageContext only has string fields here.
				// If it doesn't, our code will break.
				key := reflect.ValueOf(&pkg).Elem().FieldByName(field.Name)
				if key.IsValid() {
					if key.String() != "" {
						packageInfo = fmt.Sprintf("%s\n%s = %s", packageInfo, packageInfoKey, key.String())
					}
				}
			}
			// These are the multi-value keys, such as Requires and BuildRequires
			if field.Tag.Get("keyArray") != "" {
				// We assume that packageContext only has string array fields here.
				// If it doesn't, our code will break.
				key := reflect.ValueOf(&pkg).Elem().FieldByName(field.Name)
				if key.IsValid() {
					keyInterface := key.Interface()
					keyArray := keyInterface.([]string)

					for _, item := range keyArray {
						if packageInfoKey == "optdepend" {
							if reason, ok := pkg.Reasons[item]; ok {
								packageInfo = fmt.Sprintf("%s\n%s = %s: %s", packageInfo, packageInfoKey, item, reason)
							} else {
								packageInfo = fmt.Sprintf("%s\n%s = %s", packageInfo, packageInfoKey, item)
							}
						} else {
							packageInfo = fmt.Sprintf("%s\n%s = %s", packageInfo, packageInfoKey, item)
						}
					}
				}
			}
		}
	}

	{ // This turns version and release into a single key
		if pkg.IsSubpackage {
			packageInfo = fmt.Sprintf(
				"%s\npkgver = %s-%s",
				packageInfo,
				defaultString(pkg.Version, pkg.parentPackage.Version),
				defaultString(pkg.Release, pkg.parentPackage.Release),
			)
		} else {
			packageInfo = fmt.Sprintf("%s\npkgver = %s-%s", packageInfo, pkg.Version, pkg.Release)
		}
	}

	uname, _ := exec.Command("uname", "-m").CombinedOutput()
	unameString := strings.TrimSpace(string(uname))

	packageInfo = fmt.Sprintf("%s\narch = %s", packageInfo, unameString)

	err = ioutil.WriteFile(filepath.Join(pkgdir, ".PKGINFO"), []byte(packageInfo), 0644)
	if err != nil {
		outputError(fmt.Sprintf("Failed to generate pkginfo:\n%s", err.Error()))
	}
}

func (pkg PackageContext) GenerateCHANGELOG() {
	changelog := strings.Join(pkg.Changelog, "\n")
	if changelog != "" {
		err := ioutil.WriteFile(filepath.Join(pkg.PackageRoot(), ".CHANGELOG"), []byte(changelog), 0775)
		if err != nil {
			outputError("Failed to build changelog for package " + highlight(pkg.GetNevra()) + ": " + err.Error())
		}
		os.Chown(filepath.Join(pkg.PackageRoot(), ".CHANGELOG"), 0, 0)
	}
}

func (pkg PackageContext) GenerateINSTALL() {
	install := ""
	m := map[*[]string]string{
		&pkg.Scriptlets.PreInstall:  "pre_install",
		&pkg.Scriptlets.PostInstall: "post_install",
		&pkg.Scriptlets.PreUpgrade:  "pre_upgrade",
		&pkg.Scriptlets.PostUpgrade: "post_upgrade",
		&pkg.Scriptlets.PreRemove:   "pre_remove",
		&pkg.Scriptlets.PostRemove:  "post_remove",
	}
	for list, functionName := range m {
		if len(*list) > 0 {
			var sb strings.Builder
			sb.WriteString(functionName)
			sb.WriteString("() {\n")
			sb.WriteString(strings.Join(*list, "\n"))
			sb.WriteString("\n}\n")
			install += sb.String()
		}
	}
	if install != "" {
		err := ioutil.WriteFile(filepath.Join(pkg.PackageRoot(), ".INSTALL"), []byte(install), 0775)
		if err != nil {
			outputError("Failed to generate scriptlets for package " + highlight(pkg.GetNevra()) + ": " + err.Error())
		}
		os.Chown(filepath.Join(pkg.PackageRoot(), ".INSTALL"), 0, 0)
	}
}

func (pkg PackageContext) GenerateMTree() {
	outputStatus("Generating .MTREE for " + highlight(pkg.GetNevra()) + "...")
	home, err := os.UserHomeDir()
	if err != nil {
		outputError(fmt.Sprintf("Failed to generate mtree:\n%s", err.Error()))
	}
	if !pkg.IsSubpackage {
		os.Chdir(filepath.Join(home, "alpmbuild/package"))
	} else {
		os.Chdir(filepath.Join(home, "alpmbuild/subpackages", pkg.GetNevra()))
	}

	cmd := exec.Command("sh", "-c", `LANG=C bsdtar -c -f - --format=mtree \
	--options='!all,use-set,type,uid,gid,mode,time,size,md5,sha256,link' \
	--null --exclude .MTREE * | gzip -c -f -n > .MTREE`)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputError(fmt.Sprintf("Failed to generate mtree:\n%s", string(output)))
	}
}

func setupDirectories() error {
	if !*fakeroot {
		outputStatus("Setting up directories...")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	for _, dir := range []string{"alpmbuild/buildroot", "alpmbuild/package", "alpmbuild/sources", "alpmbuild/packages", "alpmbuild/subpackages", "alpmbuild/sourcepackages"} {
		err = os.MkdirAll(filepath.Join(home, dir), os.ModePerm)
		if dir == "alpmbuild/sourcepackages" ||
			dir == "alpmbuild/package" ||
			dir == "alpmbuild/subpackages" {
			os.RemoveAll(filepath.Join(home, dir))
			err = os.MkdirAll(filepath.Join(home, dir), os.ModePerm)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (pkg PackageContext) setupSources() error {
	if !*fakeroot {
		outputStatus("Verifiying sources...")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	handleSource := func(source Source) error {
		target := filepath.Join(home, "alpmbuild/buildroot", path.Base(source.URL))
		if source.Rename != "" {
			target = filepath.Join(home, "alpmbuild/buildroot", source.Rename)
		}
		if isValidUrl(source.URL) {
			if !*fakeroot {
				outputStatus("Downloading " + highlight(source.URL) + "...")
			}
			err := downloadFile(filepath.Join(home, "alpmbuild/sources", path.Base(source.URL)), source.URL)
			if err != nil {
				return err
			}
			if source.Rename != "" && !*fakeroot {
				outputStatus(fmt.Sprintf("Renaming %s to %s...", highlight(path.Base(source.URL)), highlight(source.Rename)))
			}
			_, err = copyFile(filepath.Join(home, "alpmbuild/sources", path.Base(source.URL)), target)
			if err != nil {
				return err
			}
			data, err := ioutil.ReadFile(filepath.Join(home, "alpmbuild/sources", path.Base(source.URL)))
			if err != nil {
				return err
			}
			badChecksum := func(filename, expected, actual string) {
				outputError(
					fmt.Sprintf(
						"Checksum failure for %s: expected %s, but got %s",
						highlight(filename),
						highlight(expected),
						highlight(actual),
					),
				)
			}
			if source.Sha1 != "" {
				if !*fakeroot {
					outputStatus("Checking sha1 integrity of " + path.Base(target) + "...")
				}
				sum := sha1.Sum(data)
				if hex.EncodeToString(sum[:]) != source.Sha1 {
					badChecksum(path.Base(target), source.Sha1, hex.EncodeToString(sum[:]))
				}
			}
			if source.Sha224 != "" {
				if !*fakeroot {
					outputStatus("Checking sha224 integrity of " + highlight(path.Base(target)) + "...")
				}
				sum := sha256.Sum224(data)
				if hex.EncodeToString(sum[:]) != source.Sha224 {
					badChecksum(path.Base(target), source.Sha1, hex.EncodeToString(sum[:]))
				}
			}
			if source.Sha256 != "" {
				if !*fakeroot {
					outputStatus("Checking sha256 integrity of " + highlight(path.Base(target)) + "...")
				}
				sum := sha256.Sum256(data)
				if hex.EncodeToString(sum[:]) != source.Sha256 {
					badChecksum(path.Base(target), source.Sha256, hex.EncodeToString(sum[:]))
				}
			}
			if source.Sha384 != "" {
				if !*fakeroot {
					outputStatus("Checking sha384 integrity of " + highlight(path.Base(target)) + "...")
				}
				sum := sha512.Sum384(data)
				if hex.EncodeToString(sum[:]) != source.Sha384 {
					badChecksum(path.Base(target), source.Sha384, hex.EncodeToString(sum[:]))
				}
			}
			if source.Sha512 != "" {
				if !*fakeroot {
					outputStatus("Checking sha512 integrity of " + highlight(path.Base(target)) + "...")
				}
				sum := sha512.Sum512(data)
				if hex.EncodeToString(sum[:]) != source.Sha512 {
					badChecksum(path.Base(target), source.Sha512, hex.EncodeToString(sum[:]))
				}
			}
			if source.Md5 != "" {
				if !*fakeroot {
					outputStatus("Checking md5 integrity of " + highlight(path.Base(target)) + "...")
				}
				sum := md5.Sum(data)
				if hex.EncodeToString(sum[:]) != source.Md5 {
					badChecksum(path.Base(target), source.Md5, hex.EncodeToString(sum[:]))
				}
			}
		} else {
			if !*fakeroot {
				outputStatus("Downloading " + highlight(source.URL) + "...")
			}
			outputStatus("Copying " + highlight(source.URL) + " to build directory...")
			if source.Rename != "" && !*fakeroot {
				outputStatus(fmt.Sprintf("Renaming %s to %s...", highlight(path.Base(source.URL)), highlight(source.Rename)))
			}
			_, err := copyFile(filepath.Join(home, "alpmbuild/sources", source.URL), filepath.Join(home, "alpmbuild/buildroot", target))
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, source := range append(pkg.Sources, pkg.Patches...) {
		err := handleSource(source)
		if err != nil {
			return err
		}
		if source.GPGSignatureURL != "" {
			err = handleSource(Source{
				URL: source.GPGSignatureURL,
			})
			if err != nil {
				return err
			}
			hasKey := func(key string) bool {
				cmd := exec.Command("gpg", "--list-keys", "0x"+key)
				cmd.Run()
				return cmd.ProcessState.ExitCode() == 0
			}
			handleSig := func(key, server string) {
				println(bold("  Do you want to try importing the key with ID ") + highlight(key) + bold(" from keyserver ") + highlight(server))
				println(bold("  [y/N]"))
				println()
				print("  " + bold("-> "))

				reader := bufio.NewReader(os.Stdin)

				text, _ := reader.ReadString('\n')

				if strings.Contains(strings.ToLower(text), "y") {
					cmd := exec.Command("gpg", "--keyserver", server, "--recv-keys", key)
					cmd.Run()
					if cmd.ProcessState.ExitCode() != 0 {
						outputWarning("Failed to import GPG key")
					}
				}

				println()
			}
			for _, keyserver := range source.GPGKeyservers {
				for _, key := range source.GPGKeys {
					if !*fakeroot {
						if !hasKey(key) {
							handleSig(key, keyserver)
						}
					}
				}
			}
			baseSource := path.Base(source.URL)
			baseSignat := path.Base(source.GPGSignatureURL)
			if !*fakeroot {
				cmd := exec.Command("gpg", "--verify", filepath.Join(home, "alpmbuild/buildroot", baseSignat), filepath.Join(home, "alpmbuild/buildroot", baseSource))
				outputStatus(
					fmt.Sprintf(
						"Verifiying of the source file %s for package %s...",
						highlight(baseSource),
						highlight(pkg.GetNevra()),
					),
				)
				cmd.Run()
				if cmd.ProcessState.ExitCode() != 0 {
					outputError(
						fmt.Sprintf(
							"Failed to verify the signature of source file %s for package %s",
							highlight(baseSource),
							highlight(pkg.GetNevra()),
						),
					)
				}
			}
		}
	}

	return nil
}

func (pkg PackageContext) PackageRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		outputError("Could not get user's home directory.")
	}

	if !pkg.IsSubpackage {
		return filepath.Join(home, "alpmbuild/package")
	}
	return filepath.Join(home, "alpmbuild/subpackages", pkg.GetNevra())
}

func (pkg PackageContext) CompressPackage() {
	outputStatus("Compressing " + highlight(pkg.GetNevra()) + " into a package...")
	home, err := os.UserHomeDir()
	if err != nil {
		outputError("Could not get user's home directory.")
	}
	packagesDir := filepath.Join(home, "alpmbuild/packages")
	os.Chdir(pkg.PackageRoot())

	clean := exec.Command("find", ".", "-type", "d", "-empty", "-delete")
	clean.Run()

	cmd := exec.Command("sh", "-c", "shopt -s dotglob; bsdtar "+CompressionTypes[*compressionType].Flag+" -cvf"+packagesDir+"/"+pkg.GetNevra()+".pkg.tar."+CompressionTypes[*compressionType].Suffix+" .PKGINFO .MTREE *")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputError(fmt.Sprintf("Creating tarball failed:\n%s", string(output)))
	}
}

func (pkg PackageContext) VerifyFiles() {
	outputStatus("Checking files of " + highlight(pkg.GetNevra()) + "...")
	home, err := os.UserHomeDir()
	if err != nil {
		outputError("Could not get user's home directory.")
	}

	pathToWalk := ""

	if !pkg.IsSubpackage {
		pathToWalk = filepath.Join(home, "alpmbuild/package")
	} else {
		pathToWalk = filepath.Join(home, "alpmbuild/subpackages", pkg.GetNevra())
	}

	err = filepath.Walk(
		pathToWalk,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				outputError("Could not verify files: " + err.Error())
			}
			if file, err := os.Stat(path); err != nil || file.Mode().IsDir() {
				return nil
			}
			truncPath := strings.TrimPrefix(
				path,
				pathToWalk,
			)
			if truncPath == "" || truncPath == "/.MTREE" || truncPath == "/.PKGINFO" || truncPath == "/.CHANGELOG" || truncPath == "/.INSTALL" {
				return nil
			}
			hasMatch := false
			for _, listedFile := range pkg.Files {
				regexString := strings.ReplaceAll(listedFile, "/", "\\/")
				regexString = strings.ReplaceAll(listedFile, ".", "\\.")
				regexString = strings.ReplaceAll(regexString, "*", ".*")
				regex, err := regexp.Compile(regexString)
				if err != nil {
					outputError("Malformed files listing: " + err.Error())
				}
				if regex.MatchString(truncPath) {
					hasMatch = true
				}
			}
			for _, subpackage := range pkg.Subpackages {
				for _, listedFile := range subpackage.Files {
					regexString := strings.ReplaceAll(listedFile, "/", "\\/")
					regexString = strings.ReplaceAll(listedFile, ".", "\\.")
					regexString = strings.ReplaceAll(regexString, "*", ".*")
					regex, err := regexp.Compile(regexString)
					if err != nil {
						outputError("Malformed files listing: " + err.Error())
					}
					if regex.MatchString(truncPath) {
						hasMatch = true
					}
				}
			}
			if !hasMatch {
				outputError("File not listed:\t" + truncPath)
			}
			return nil
		},
	)

	if err != nil {
		outputError("Could not verify files: " + err.Error())
	}
}

func (pkg PackageContext) ClearTimestamps() {
	outputStatus("Cleaning up timestamps...")
	home, err := os.UserHomeDir()
	if err != nil {
		outputError("Could not get user's home directory.")
	}
	path := filepath.Join(home, "alpmbuild/package")
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		cmd := exec.Command("touch", "-d", "@1", path)
		cmd.Run()
		return nil
	})
}

func (pkg PackageContext) TakeFilesFromParent() {
	outputStatus("Moving files from " + highlight(pkg.parentPackage.Name) + " to " + highlight(pkg.GetNevra()) + "...")
	home, err := os.UserHomeDir()
	if err != nil {
		outputError("Could not get user's home directory.")
	}
	path := filepath.Join(home, "alpmbuild/subpackages", pkg.GetNevra())
	os.MkdirAll(path, os.ModePerm)

	for _, file := range pkg.Files {
		globPath := filepath.Join(home, "alpmbuild/package", file)
		files, err := filepath.Glob(globPath)
		if err != nil {
			outputError("Bad globbing: " + err.Error())
		}
		for _, fileToCopy := range files {
			fromPackageRootPath := strings.TrimPrefix(fileToCopy, filepath.Join(home, "alpmbuild/package"))
			dirName, _ := filepath.Split(fromPackageRootPath)
			os.MkdirAll(filepath.Join(path, dirName), os.ModePerm)
			err = os.Rename(fileToCopy, filepath.Join(path, fromPackageRootPath))
			if err != nil {
				outputError("Failed to take file from parent package: " + err.Error())
			}
		}
	}
}

func (pkg PackageContext) GenerateSourcePackage() {
	outputStatus("Generating source package...")
	home, err := os.UserHomeDir()
	if err != nil {
		outputError("There was an error obtaining the home directory")
	}
	for _, source := range pkg.Sources {
		if !isValidUrl(source.URL) {
			_, err := copyFile(filepath.Join(home, "alpmbuild/sources", source.URL), filepath.Join(home, "alpmbuild/sourcepackages", source.URL))
			if err != nil {
				outputError("There was an error copying sources into the source package")
			}
		}
	}
	os.Chdir(startPWD)
	_, err = copyFile(*buildFile, filepath.Join(home, "alpmbuild/sourcepackages", path.Base(*buildFile)))
	if err != nil {
		outputError("There was an error copying the specfile into the source package:\n\t" + err.Error())
	}
	os.Chdir(filepath.Join(home, "alpmbuild"))
	err = os.RemoveAll(filepath.Join(home, "alpmbuild", pkg.GetNevr()))
	if err != nil {
		outputError("Failed to clean up source package directory: " + err.Error())
	}
	err = os.Rename(filepath.Join(home, "alpmbuild/sourcepackages"), filepath.Join(home, "alpmbuild", pkg.GetNevr()))
	if err != nil {
		outputError("Failed to rename source package directory: " + err.Error())
	}
	err = exec.Command("bsdtar", CompressionTypes[*compressionType].Flag, "-cvf", filepath.Join(home, "alpmbuild", "packages", pkg.GetNevr()+".alpmsrc.pkg.tar."+CompressionTypes[*compressionType].Suffix), pkg.GetNevr()).Run()
	if err != nil {
		outputError("Failed to compress source package: " + err.Error())
	}
	err = os.RemoveAll(filepath.Join(home, "alpmbuild", pkg.GetNevr()))
	if err != nil {
		outputError("Failed to clean up source package directory: " + err.Error())
	}
	outputStatus("Generated source package")
}

func (pkg *PackageContext) InheritFromParent() {
	if pkg.IsSubpackage {
		if pkg.Epoch == "" {
			pkg.Epoch = pkg.parentPackage.Epoch
		}
		if pkg.Version == "" {
			pkg.Version = pkg.parentPackage.Version
		}
		if pkg.Release == "" {
			pkg.Release = pkg.parentPackage.Release
		}
		if pkg.License == "" {
			pkg.License = pkg.parentPackage.License
		}
	}
}

func (pkg PackageContext) CheckArch() {
	uname, _ := exec.Command("uname", "-m").CombinedOutput()
	unameString := strings.TrimSpace(string(uname))
	if len(pkg.ExclusiveArch) > 0 {
		for _, arch := range pkg.ExclusiveArch {
			if arch == unameString {
				return
			}
		}
	} else {
		return
	}
	outputError(
		fmt.Sprintf(
			"System architecture %s is not in the list of available arches for %s: %s",
			highlight(unameString),
			highlight(pkg.GetNevr()),
			strings.Join(pkg.ExclusiveArch, ", "),
		),
	)
}

func (pkg PackageContext) BuildPackage() {
	pkg.CheckArch()
	if !*fakeroot {
		outputStatus("Building package " + highlight(pkg.GetNevra()) + "...")
	}
	err := setupDirectories()
	if err != nil {
		outputError(fmt.Sprintf("Error setting up directories:\n\t%s", err.Error()))
	}
	err = pkg.setupSources()
	if err != nil {
		outputError(fmt.Sprintf("Error setting up sources:\n\t%s", err.Error()))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		outputError("Could not get user's home directory.")
	}

	os.Chdir(filepath.Join(home, "alpmbuild/buildroot"))

	env := os.Environ()
	env = append(env, fmt.Sprintf("BUILDROOT=%s", filepath.Join(home, "alpmbuild/package")))

	// Prepare commands.
	var commands []string
	var installCommands []string

	commands = append(commands, pkg.Commands.Prepare...)
	commands = append(commands, pkg.Commands.Build...)
	commands = append(commands, pkg.Commands.Check...)
	installCommands = append(installCommands, pkg.Commands.Install...)

	path, err := writeTempfile(strings.Join(commands, "\n"))
	if err != nil {
		outputError("There was an error preparing a temporary file.")
	}

	installPath, err := writeTempfile(strings.Join(installCommands, "\n"))
	if err != nil {
		outputError("There was an error preparing a temporary file.")
	}

	var pathToUse string
	if *fakeroot {
		pathToUse = installPath
	} else {
		pathToUse = path
	}

	cmd := exec.Command("sh", pathToUse)
	cmd.Env = env

	if !*hideCommandOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if !*fakeroot {
		outputStatus("Building...")
	}
	err = cmd.Run()
	if err != nil {
		outputError("Exit status was non-zero in build script, aborting...")
	}

	if !*fakeroot {
		os.Chdir(initialWorking)
		cmd := exec.Command("fakeroot", append(os.Args, "-fakeroot")...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	outputStatus("Running package commands...")

	pkg.ClearTimestamps()

	for _, subpackage := range pkg.Subpackages {
		subpackage.InheritFromParent()
		subpackage.TakeFilesFromParent()
		subpackage.lintAll()
		subpackage.GenerateINSTALL()
		subpackage.GenerateCHANGELOG()
		subpackage.GeneratePackageInfo()
		subpackage.GenerateMTree()
		subpackage.CompressPackage()
		subpackage.VerifyFiles()
	}

	pkg.lintAll()
	pkg.GenerateINSTALL()
	pkg.GenerateCHANGELOG()
	pkg.GeneratePackageInfo()
	pkg.GenerateMTree()
	pkg.ClearTimestamps()
	pkg.CompressPackage()
	if *checkFiles {
		pkg.VerifyFiles()
	}
	if *generateSourcePackage {
		pkg.GenerateSourcePackage()
	}
}
