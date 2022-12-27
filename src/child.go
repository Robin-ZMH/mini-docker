package src

func Child(args ...string) {
	imageHash, containerID := args[0], args[1]
	executeContainer(imageHash, containerID, args[2:]...)
}
