BSON_PKGS = $(shell ./etc/find_pkgs.sh ./bson)
BSON_TEST_PKGS = $(shell ./etc/find_pkgs.sh ./bson _test)
PKGS = $(BSON_PKGS)
TEST_PKGS = $(BSON_TEST_PKGS)

.PHONY: default
default: check-fmt vet lint errcheck

.PHONY: check-fmt
check-fmt:
	@gofmt -l -s $(PKGS) | read; if [ $$? == 0 ]; then echo "gofmt check failed for:"; gofmt -l -s $(PKGS) | sed -e 's/^/ - /'; exit 1; fi

.PHONY: fmt
fmt:
	gofmt -l -s -w $(PKGS)

.PHONY: lint
lint:
	golint $(PKGS) | ./etc/lintscreen.pl .lint-whitelist

.PHONY: lint-add-whitelist
lint-add-whitelist:
	golint $(PKGS) | ./etc/lintscreen.pl -u .lint-whitelist
	sort .lint-whitelist -o .lint-whitelist

.PHONY: errcheck
errcheck:
	errcheck ./bson/...

.PHONY: vet
vet:
	go tool vet -cgocall=false -composites=false -structtags=false -unusedstringmethods="Error" $(PKGS)
