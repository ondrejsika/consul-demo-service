VERSION = 1
IMAGE = ondrejsika/consul-demo-service:$(VERSION)
CONTAINER = consul-demo-service

all: build push

build:
	docker build -t $(IMAGE) .

push:
	docker push $(IMAGE)

up:
	docker run -d --name $(CONTAINER) -p 80:80 $(IMAGE)

down:
	docker rm -f $(CONTAINER)

logs:
	docker logs -f $(CONTAINER)
