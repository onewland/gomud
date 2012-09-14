all:
	go install mud
	go install mud/simple
	go build
	go test mud
	go test

deploy:
	rsync -avz  \
		--exclude "pkg/*" \
		--exclude "*~"  \
		--recursive -t ./** prgmr:gomud

start:
	ssh prgmr "nohup ./gomud/gomud &> ./gomud/gomud.log &"