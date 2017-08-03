PACKAGE = github.com/r11t/check-break/src


default : build

build:
	go build -o check-break -v $(PACKAGE)/...

major:
	@semver inc major

minor:
	@semver inc minor

patch:
	@semver inc patch
