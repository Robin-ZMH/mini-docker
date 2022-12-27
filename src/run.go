package src

import (
	"log"
	"os"
)

func Run(args ...string) {
	containerID := genRandID()
	log.Printf("New container ID: %s\n", containerID)

	imageHash := initImage(args[0])
	cmdArgs := args[1:]

	log.Printf("Image to start: %s\n", imageHash)

	must_ok(createContainerDirectories(containerID))
	log.Println("successfully create Container directories")

	must_ok(mountOverlayFileSystem(containerID, imageHash))
	log.Println("successfully mount Overlay-File-systems")

	log.Printf("Container start...")
	must_ok(executeChildCMD(imageHash, containerID, cmdArgs...))
	

	//clean up
	log.Println("clean up...")
	must_ok(unmountContainerFs(containerID))
	must_ok(os.RemoveAll(getPathOfContainer(containerID)))
}
