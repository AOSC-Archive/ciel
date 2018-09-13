package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"ciel/display"
	"ciel/internal/ciel"
	"ciel/internal/container/instance"
	"ciel/systemd-api/nspawn"
)

const (
	LatestTarballURL = "https://repo.aosc.io/aosc-os/os-amd64/buildkit/aosc-os_buildkit_latest_amd64.tar.xz"
)

func untarGuestOS() {
	basePath := flagCielDir()
	batchFlag := flagBatch()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	tar := flag.Arg(0)
	if tar == "" {
		d.SECTION("Download OS")
		d.ITEM("latest tarball url")
		d.Println(d.C(d.CYAN, LatestTarballURL))
		tarUrl, _ := url.Parse(LatestTarballURL)
		_, DownloadTarball := filepath.Split(tarUrl.Path)
		if DownloadTarball == "" {
			DownloadTarball = "latest.tar.xz"
		}
		cmd := exec.Command("curl", "-o", DownloadTarball, "-#", LatestTarballURL)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		d.ITEM("download")
		if err != nil {
			d.FAILED_BECAUSE(err.Error())
			os.Remove(DownloadTarball)
			return
		}
		d.OK()
		tar = DownloadTarball
	}

	d.SECTION("Load OS From Compressed File")
	d.ITEM("are there any instances?")

	if instList := c.GetAllNames(); len(instList) != 0 {
		d.Println(d.C(d.YELLOW, strings.Join(instList, " ")))
		if !*batchFlag && d.ASKLower("DELETE ALL INSTANCES?", "yes/no") != "yes" {
			os.Exit(1)
		}
		for _, inst := range c.GetAll() {
			if inst.Running() {
				inst.Stop(context.TODO())
			}
			if inst.Mounted() {
				inst.Unmount()
			}
			d.ITEM("delete " + inst.Name)
			err := c.DelInst(inst.Name)
			d.ERR(err)
		}
	} else {
		d.Println(d.C(d.CYAN, "NO"))
	}

	d.ITEM("is dist dir empty?")
	os.Mkdir(c.DistDir(), 0755)
	list, err := ioutil.ReadDir(c.DistDir())
	if len(list) != 0 {
		d.Println(d.C(d.YELLOW, "NO"))
		if !*batchFlag && d.ASKLower("DELETE the old OS?", "yes/no") != "yes" {
			os.Exit(1)
		}
		d.ITEM("remove dist dir")
		if err := os.RemoveAll(c.DistDir()); err != nil {
			d.FAILED_BECAUSE(err.Error())
			os.Exit(1)
		}
		d.OK()

		d.ITEM("re-create dist dir")
		if err := os.Mkdir(c.DistDir(), 0755); err != nil {
			d.FAILED_BECAUSE(err.Error())
			os.Exit(1)
		}
		d.OK()
	} else if err != nil {
		d.FAILED_BECAUSE(err.Error())
		os.Exit(1)
	} else {
		d.Println(d.C(d.CYAN, "YES"))
	}

	d.ITEM("unpacking os...")
	cmd := exec.Command("tar", "-xpf", tar, "-C", c.DistDir())
	output, err := cmd.CombinedOutput()
	if err != nil {
		d.FAILED_BECAUSE(strings.TrimSpace(string(output)))
	}
	d.OK()
}

func update() {
	var runErr error
	var exitStatus int
	defer func() {
		if runErr != nil {
			os.Exit(1)
		} else if exitStatus != 0 {
			os.Exit(exitStatus)
		}
	}()

	basePath := flagCielDir()
	networkFlag := flagNetwork()
	batchFlag := flagBatch()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()

	d.SECTION("Update Guest Operating System")
	d.ITEM("are there online instances?")
	ready := true
	for _, inst := range c.GetAll() {
		if inst.Running() || inst.Mounted() {
			ready = false
			d.Print(d.C(d.YELLOW, inst.Name) + " ")
		}
	}
	if ready {
		d.Print(d.C(d.CYAN, "NO"))
	}
	d.Println()

	if !ready {
		if !*batchFlag && d.ASKLower("Stop all instances?", "yes/no") != "yes" {
			os.Exit(1)
		}
		for _, inst := range c.GetAll() {
			if inst.Running() {
				inst.Stop(context.TODO())
			}
			if inst.Mounted() {
				inst.Unmount()
			}
		}
	}

	const instName = "__update__"
	d.ITEM("create temporary instance")
	c.AddInst(instName)
	d.OK()
	defer func() {
		d.ITEM("delete temporary instance")
		c.DelInst(instName)
		d.OK()
	}()
	inst := c.Instance(instName)
	d.ITEM("mount temporary instance")
	inst.Mount()
	d.OK()
	defer func() {
		inst.Unmount()
	}()

	type ExitError struct{}
	var run = func(cmd string) (int, error) {
		return _shellRun(inst, *networkFlag, true, cmd)
	}
	defer func() {
		p := recover()
		if p == nil {
			return
		}
		if _, isExit := p.(ExitError); !isExit {
			panic(p)
		}
		if runErr != nil {
			d.Println(d.C(d.RED, runErr.Error()))
		} else {
			d.Println(d.C(d.YELLOW, "INTERRUPTED, exit status: "+strconv.Itoa(exitStatus)))
		}
	}()

	exitStatus, runErr = run(`apt update --yes`)
	d.ITEM("update database")
	if runErr != nil || exitStatus != 0 {
		panic(ExitError{})
	}
	d.OK()

	exitStatus, runErr = run(`apt -o Dpkg::Options::="--force-confnew" full-upgrade --autoremove --purge --yes`)
	d.ITEM("update and auto-remove packages")
	if runErr != nil || exitStatus != 0 {
		panic(ExitError{})
	}
	d.OK()

	d.ITEM("merge changes")
	err := inst.FileSystem().Merge()
	d.ERR(err)
}

func factoryReset() {
	var runErr error
	var exitStatus int
	defer func() {
		if runErr != nil {
			os.Exit(1)
		} else if exitStatus != 0 {
			os.Exit(exitStatus)
		}
	}()

	basePath := flagCielDir()
	instName := flagInstance()
	parse()

	i := &ciel.Ciel{BasePath: *basePath}
	i.Check()
	c := i.Container()
	c.CheckInst(*instName)
	inst := c.Instance(*instName)

	d.SECTION("Factory Reset Guest Operating System")

	inst.Stop(context.TODO())

	ctnInfo := buildContainerInfo(false, false)
	runInfo := buildRunInfo([]string{
		"/bin/apt-gen-list",
	})
	if exitStatus, err := inst.Run(context.TODO(), ctnInfo, runInfo); exitStatus != 0 {
		log.Println(err)
	}

	inst.Stop(context.TODO())
	d.ITEM("mount instance")
	inst.Mount()
	d.OK()

	d.ITEM("collect package list in dpkg")
	pkgList := dpkgPackages(inst)
	d.OK()

	d.ITEM("collect file set in packages")
	fileSet := dpkgPackageFiles(inst, pkgList)
	d.OK()

	i.GetTree().MountHandler(inst, false)

	d.ITEM("remove out-of-package files")
	err := clean(inst.MountPoint(), fileSet,
		[]string{
			`^/tree`,
			`^/dev`,
			`^/efi`,
			`^/etc`,
			`^/run`,
			`^/usr`,
			`^/var/lib/apt/gen/status\.json$`,
			`^/var/lib/apt/extended_states`,
			`^/var/lib/dpkg`,
			`^/var/log/journal$`,
			`^/root`,
			`^/home`,
			`/\.updated$`,
		}, []string{
			`^/etc/.*-$`,
			`^/etc/machine-id`,
			`^/etc/ssh/ssh_host_.*`,
			`^/root/.bash_history`,
			`^/var/lib/dpkg/[^/]*-old`,
			`^/var/tmp/*`,
			`^/var/log/apt/*`,
			`^/var/log/alternative.log`,
			`^/var/log/journal/^(?!remote).*`,
		}, func(path string, info os.FileInfo, err error) error {
			if err := os.RemoveAll(path); err != nil {
				log.Println("clean:", err.Error())
			}
			return nil
		})
	d.ERR(err)
}

func clean(root string, packageFiles map[string]bool, preserve []string, delete []string, fn filepath.WalkFunc) error {
	var preserveList []string
	if preserve != nil {
		for _, re := range preserve {
			preserveList = append(preserveList, "("+re+")")
		}
	} else {
		preserveList = []string{`($^)`}
	}
	var deleteList []string
	if delete != nil {
		for _, re := range delete {
			deleteList = append(deleteList, "("+re+")")
		}
	} else {
		deleteList = []string{`($^)`}
	}
	regexPreserve := regexp.MustCompile("(" + strings.Join(preserveList, "|") + ")")
	regexDelete := regexp.MustCompile("(" + strings.Join(deleteList, "|") + ")")

	if packageFiles == nil {
		return errors.New("no file in dpkg")
	}
	filepath.Walk(root, wrapWalkFunc(root, func(path string, info os.FileInfo, err error) error {
		if _, inDpkg := packageFiles[path]; inDpkg {
			return nil
		}
		if regexDelete.MatchString(path) {
			return fn(filepath.Join(root, path), info, err)
		}
		if regexPreserve.MatchString(path) {
			return nil
		}
		return fn(filepath.Join(root, path), info, err)
	}))

	return nil
}

func wrapWalkFunc(root string, fn filepath.WalkFunc) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.Join("/", rel)
		return fn(rel, info, err)
	}
}

func dpkgPackages(i *instance.Instance) []string {
	ctnInfo := buildContainerInfo(false, false)

	stdout := new(bytes.Buffer)
	var args []string
	args = []string{
		"/usr/bin/dpkg-query",
		"--show",
		"--showformat=${Package}\n",
	}
	runInfo := buildRunInfo(args)
	runInfo.StdDev = &nspawn.StdDevInfo{
		Stderr: os.Stderr,
		Stdout: stdout,
	}
	if exitStatus, err := i.Run(context.TODO(), ctnInfo, runInfo); exitStatus != 0 {
		log.Println(err)
		return nil
	}
	pkgList := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	for i := range pkgList {
		pkgList[i] = strings.TrimSpace(pkgList[i])
	}
	return pkgList
}

func dpkgPackageFiles(i *instance.Instance, packages []string) map[string]bool {
	ctnInfo := buildContainerInfo(false, false)

	stdout := new(bytes.Buffer)
	var args []string
	args = []string{
		"/usr/bin/dpkg-query",
		"--listfiles",
	}
	args = append(args, packages...)
	runInfo := buildRunInfo(args)
	runInfo.StdDev = &nspawn.StdDevInfo{
		Stderr: os.Stderr,
		Stdout: stdout,
	}
	if exitStatus, err := i.Run(context.TODO(), ctnInfo, runInfo); exitStatus != 0 {
		log.Println(err)
		return nil
	}

	hashMap := make(map[string]bool, 100000)
	dataSet := strings.Split(stdout.String(), "\n")
	root := i.MountPoint()
	for _, record := range dataSet {
		record = strings.TrimSpace(record)
		if len(record) == 0 {
			continue
		}
		path := strings.TrimSpace(record)
		evalSymlinksCleanCache()
		path = evalSymlinks(root, path, true)
		hashMap[path] = true
	}
	return hashMap
}

var cachedLstat map[string]bool

// evalSymlinks resolves symbolic links IN path based on specified root, outputs an unique path.
//
// noLeaf: true - do not resolve the last object in path (file or directory). true in common cases.
func evalSymlinks(root string, path string, noLeaf bool) string {
	path = filepath.Clean(path)
	var pos int
	var prefix string
	for pos != len(path) {
		// split
		if delimPos := strings.IndexRune(path[pos:], filepath.Separator); delimPos == -1 {
			// deepest one
			if noLeaf {
				break
			}
			pos = len(path)
			prefix = path
		} else {
			// directories in the middle
			pos += delimPos + 1
			if pos == 1 {
				prefix = path[:pos]
			} else {
				prefix = path[:pos-1]
			}
		}

		isLink, cached := cachedLstat[prefix]
		if cached {
			if !isLink {
				continue // most common
			}
			panic("loop in symlink")
		} else {
			fi, _ := os.Lstat(filepath.Join(root, prefix))
			if fi == nil || fi.Mode()&os.ModeSymlink == 0 {
				cachedLstat[prefix] = false
				continue
			}
			cachedLstat[prefix] = true
			target, _ := os.Readlink(filepath.Join(root, prefix))
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(prefix), target)
			}
			target = evalSymlinks(root, target, noLeaf) // fix new path (prefix)
			path = filepath.Join(target, path[pos:])
			pos = len(target)
		}
	}
	return path
}

func evalSymlinksCleanCache() {
	cachedLstat = make(map[string]bool)
}
