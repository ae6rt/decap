DOCKER_USER := ae6rt

# No changes required below this line ------------

IMAGE := $(DOCKER_USER)/java7:latest

container:
	docker build -t $(IMAGE) .  

push: container
	docker push $(IMAGE)
