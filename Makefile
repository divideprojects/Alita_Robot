test:
	@pre-commit run --all-files

install:
	@python3 -m pip install --upgrade pip setuptools
	@python3 -m pip install -r requirements.txt

run:
	@python3 -m alita

update:
	@git pull
	@python3 -m pip install --upgrade -r requirements.txt
