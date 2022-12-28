package src

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func createContainerDirectories(containerID string) error {
	path := getPathOfContainer(containerID)
	dirs := []string{path + "/fs", path + "/fs/mnt", path + "/fs/upperdir", path + "/fs/workdir"}
	return createDirs(dirs)
}

func mountOverlayFileSystem(containerID, imageHash string) error {
	layers := getLayersOfImage(imageHash)
	containerFS := getFSHomeOfContainer(containerID)
	// mntOptions := "lowerdir=" + strings.Join(layers, ":") + ",upperdir=" + containerFS + "/upperdir,workdir=" + containerFS + "/workdir"
	mntOptions := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		strings.Join(layers, ":"),
		containerFS+"/upperdir",
		containerFS+"/workdir")
	return syscall.Mount("none", getMountPathOfContainer(containerID), "overlay", 0, mntOptions)
}

func unmountContainerFs(containerID string) error {
	mountPath := getMountPathOfContainer(containerID)
	return syscall.Unmount(mountPath, 0)
}

func executeChildCMD(imageHash, containerID string, args ...string) error {
	args = append([]string{"child", imageHash, containerID}, args...)
	cmd := exec.Cmd{
		Path:   "/proc/self/exe",
		Args:   append([]string{"/proc/self/exe"}, args...),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		SysProcAttr: &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWPID |
				syscall.CLONE_NEWNS |
				syscall.CLONE_NEWUTS |
				syscall.CLONE_NEWIPC,
		},
	}
	return cmd.Run()
}

/*
executeContainer will finally execute the container start command after some initialazition
*/
func executeContainer(imageHash, containerID string, args ...string) {
	conf := readContainerConfig(imageHash)
	mnt := getMountPathOfContainer(containerID)

	cmd := exec.Cmd{
		Path:   args[0],
		Args:   args[1:],
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Env:    conf.Config.Env,
	}

	syscall.Sethostname([]byte(containerID))
	syscall.Chroot(mnt)
	syscall.Chdir("/")
	must_ok(createDirs([]string{"/proc", "/sys"}))
	must_ok(syscall.Mount("proc", "/proc", "proc", 0, ""))
	must_ok(syscall.Mount("sysfs", "/sys", "sysfs", 0, ""))

	cmd.Run()

	must_ok(syscall.Unmount("/sys", 0))
	must_ok(syscall.Unmount("/proc", 0))
}
