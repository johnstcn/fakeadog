all: docker fakeadog

.PHONY: docker

docker:
	docker build -t johnstcn/fakeadog .

fakeadog:
	CGO_ENABLED=0 GOOS=linux go build  -ldflags '-w' github.com/johnstcn/fakeadog/cmd/fakeadog

clean:
	docker rmi johnstcn/fakeadog || true
	rm -f fakeadog || true
