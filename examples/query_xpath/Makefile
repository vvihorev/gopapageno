generate:
	# gopapageno -l query_xpath.l -g query_xpath.g -types-only -log
	gopapageno -l query_xpath.l -g query_xpath.g -log
	go build .

test:
	python test_driver.py

clean:
	rm -rf *.pg.go query_xpath
