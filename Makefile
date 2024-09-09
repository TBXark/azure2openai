.PHONY: buildLinuxX86
buildLinuxX86:
	GOOS=linux GOARCH=amd64 go build -o ./build/linux_x86/ ./...

.PHONY: deploy
deploy:
	ansible-playbook -i inventory.ini playbook.yaml