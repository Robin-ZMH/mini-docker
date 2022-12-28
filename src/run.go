package src

import (
	"log"
	"os"
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

func clean_up(containerID string) {
	log.Println("Clean up...")
	_ = unmountContainerFs(containerID)
	_ = os.RemoveAll(getHomeOfContainers())
	log.Println("All temporary container files are removed")
}
