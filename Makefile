all:
	go install mud
	go install mud/simple
	go build

deploy:
	rsync -avz  \
		--exclude "pkg/*" \
		--exclude "*~"  \
		--recursive -t ./** prgmr:gomud

start:
	ssh prgmr "nohup ./gomud/gomud &> ./gomud/gomud.log &"