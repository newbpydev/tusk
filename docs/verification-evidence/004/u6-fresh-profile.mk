fresh-profile:
	python3 /tmp/tusk-fresh-profile.py
	go tool pprof -top -cum -nodecount=55 /tmp/tusk-u6-fresh-profiles/sample-*.cpu
	go tool pprof -top -nodecount=30 /tmp/tusk-u6-fresh-profiles/sample-*.cpu
