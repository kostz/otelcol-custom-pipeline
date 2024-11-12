VERSION ?= 0.110.0
DEF_BUILDER = ./ocb
BUILDER ?= $(DEF_BUILDER)

clean:
	rm -rf ./otel
	rm -rf ./ocb
	mkdir otel

ocb:
	if [ "$(BUILDER)" = "$(DEF_BUILDER)" ]; then \
		[ -x $(BUILDER) ] && exit 0 ; \
		sys=$$( uname -s | tr 'A-Z' 'a-z' ); \
		mach=$$( uname -m | tr 'A-Z' 'a-z' ); \
		[ $$mach = "x86_64" ] && mach="amd64"; \
		curl -o ocb -L "https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/cmd/builder/v$(VERSION)/ocb_$(VERSION)_$${sys}_$${mach}" ;\
		chmod 777 ./ocb; \
	fi

build: ocb
	CGO_ENABLED=0 $(BUILDER) --config config_build.yaml --verbose