IMAGE=ae6rt/aftomato-build-base:latest

container:
	docker build -t $(IMAGE) .  

push: container
	docker push $(IMAGE)
