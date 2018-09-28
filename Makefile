
FILES = $(shell find . -type f -name '*.go' -not -path './vendor/*')

gofmt:
	@gofmt -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)

deps:
	go get -u github.com/mgechev/revive

	go get -u github.com/altipla-consulting/sentry
	go get -u github.com/julienschmidt/httprouter
	go get -u github.com/sirupsen/logrus

test:
	revive -formatter friendly
	go install .
