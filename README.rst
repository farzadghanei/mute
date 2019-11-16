****
Mute
****

`mute` runs others programs while suppressing the output under configured
conditions. A good use case is to keep cron jobs silenced and avoid receiving
emails for known conditions.

It's written in Go, has a small resource overhead with no runtime dependencies.


Usage
=====

.. code-block::

    # by default exit code "0" is muted
    mute bash -c "echo 'this is muted'"
    mute bash -c "echo 'this is printed, exiting with 12'; exit 12"


`mute` accepts a command with optional arguments to run. `mute` itself
has no arguments and can be configured with a file (in `TOML <https://github.com/toml-lang/toml>`_),
and environment variables.

Configuration is validated before running the program.

The exit code of `mute` is the exit code of the command it runs.
However `mute` exits with 127 (`mute.ExitErrExec`) when failed to execute the commnad,
and with 126 (`mute.ExitErrConf`) when configuration is invalid.


Configuration
-------------

If the configuration file does not exist, or is not accessile (permissions, etc.)
mute continues with the default configuration.
Any accessible configuration should be valid otherwise mute exits with `mute.ExitErrConf`.


Default Config
==============
When there is no config specified, mute suppresses output from successful runs, matching
exit code 0 and any output pattern.


Environment Variables
=====================

  * `MUTE_CONFIG`: full/relative path to the config file. default is `/etc/mute.toml`, no file no issue.
    an empty value means no config file lookup.


Configuration File
===================

The accessible configuration file should contain valid criteria defenitions in TOML format.


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


License
=======

`mute` is an open source project released under the terms of the MIT license.

The MIT License (MIT)

Copyright (c) 2019 Farzad Ghanei

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
