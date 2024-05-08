# The `.PHONY` target is used to specify that the following target names are not actual files.
# It ensures that Make doesn't get confused by file names that match target names.
.PHONY: setup

# The `setup` target is used to run the setup script.
# It is a command that should be run to configure the system or install necessary dependencies
# for the project to run properly.
setup:
# Executes the setup script located in the `scripts` directory.
# This script includes all necessary installation commands and configuration settings.
	scripts/setup.sh
