.PHONY: buildLinuxX86
buildLinuxX86:
	GOOS=linux GOARCH=amd64 go build -o ./build/linux_x86/ ./...

.PHONY: deploy
deploy:
	ansible-playbook -i inventory.ini playbook.yaml


.PHONY: buildImage
buildImage: buildLinuxX86
	docker buildx build --platform=linux/amd64 -t ghcr.io/tbxark/azure2openai:latest .
	docker push ghcr.io/tbxark/azure2openai:latest