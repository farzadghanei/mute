====
MUTE
====

---------------------------------------------------------------------
runs other programs and prevents the output under configured criteria
---------------------------------------------------------------------

:Author: Farzad Ghanei <farzad.ghanei@tutanota.com>
:Date:   2020-02-16
:Copyright:  Copyright (c) 2019 Farzad Ghanei. mute is an open source project released under the terms of the MIT license.
:Version: 0.2.0
:Manual section: 1
:Manual group: General Command Manuals


SYNOPSIS
========
    mute COMMAND [COMMAND OPTIONS]

DESCRIPTION
===========
mute accepts a command with optional arguments to run. mute itself takes no arguments
but can be configured with a file, and environment variables.
The configuration is validated before running the command.

A good use case is to keep cron jobs silenced and avoid receiving emails for known conditions.

mute matches the exit code and the standard output of the command it runs against a set of criteria,
and when it finds a match discards the output.
Each criteria is a list of exit codes, and one or more regular expression patterns (matching stdout).

OPTIONS
===========
mute does not accept any options.

EXIT STATUS
===========
The exit code of mute is the exit code of the command it runs. However mute exits with:

**127**: when failed to execute the commnad

**126**: when configuration is invalid

ENVIRONMENT
===========
mute can be configured with these environment variables:

**MUTE_EXIT_CODES**: comma separated list of exit codes to mute (same as **exit_codes** in **mute.default** config)

**MUTE_STDOUT_PATTERN**: regex pattern to mute the output when stdout matches

**MUTE_CONFIG**: absolute/relative path to the config file. default is /etc/mute.toml, no file no issue.
an empty value means no config file lookup.

If the criteria is defined via environment variables (**MUTE_EXIT_CODES**, **MUTE_STDOUT_PATTERN**), configuration file
is not checked at all.


FILES
=====

**\/etc\/mute.toml**
    The default configuration file, if available should contain valid criteria defenitions in TOML format.
    The path to this file can be set by **MUTE_CONFIG** environment variable.


Example configuration


.. code-block::

    # When a command matched this criteria, the output will be muted.
    # Exit codes and stdout patterns are grouped by "AND", requiring both to match.
    # Multiple sections will be grouped by "OR", so matching any section will suppress the output.
    # stdout is checked by matching with regular expression patterns.

    [[ default ]]
    exit_codes = [0]  # any of exit codes could match

    # OR
    [[ default ]]
    stdout_patterns = [".+ OK .+"]  # stdout matches any listed regex patterns

    # OR
    [[ default ]]
    exit_codes = [1, 2]  # any program that exits with either 1,2 AND prints OK
    stdout_patterns = ["OK"]

    [ commands ]
    # Command specific settings, overriding default settings, not stacking with default.
    # This applies to any command starting with 'user': 'user' and 'useradd' and 'userdel'

      [[ commands.user ]]
      exit_codes = [0]  # any command starting with "user" will match ONLY when exit code is 0

      # Command specific settings can also be grouped with OR by repeating the settings
      [[ commands.user ]]
      stdout_patterns = ["^$"]  # now any command starting with "user" will match when output is empty regardless of exit code


REPORTING BUGS
==============
Bugs can be reported with https://github.com/farzadghanei/mute/issues
