package src

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func Run(args ...string) {
	cmdArgs := args[1:]
	containerID := genRandID()
	defer clean_up(containerID)
	log.Printf("New container ID: %s\n", containerID)

	imageHash := initImage(args[0])
	log.Printf("Image to start: %s\n", imageHash)

	must_ok(createContainerDirectories(containerID))
	log.Println("Successfully create Container directories")

	must_ok(mountOverlayFileSystem(containerID, imageHash))
	log.Println("Successfully mount Overlay-File-systems")

	/* because go can't:
	1. sethostname of child process
	2. chroot of of child process
	3. change the directory of child process
	so, we can fork a child process to re-execute the program's exe,
	in the child process, we implement the above things
	and then fork a process again to execute the container command.
	*/
	log.Printf("Container start...")
	must_ok(executeChildCMD(imageHash, containerID, cmdArgs...))
}

func Child(args ...string) {
	imageHash, containerID := args[0], args[1]
	executeContainer(imageHash, containerID, args[2:]...)
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
