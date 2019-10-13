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

The exit code of `mute` is the exit code of the command to execute.
`mute` exits with 127 (`mute.exec.EXEC_ERR`) when failed to execute the commnad.


Environment Variables
---------------------

Environment variables override config file directives.

  * `MUTE_CONFIG`: full/relative path to the config file. default is `/etc/mute.toml`, no file no issue.
  * `MUTE_EXIT_CODES`: comma separated list of exit codes to suppress (same as `exit_codes` in `mute.default` config)
  * `MUTE_STDOUT_PATTERN`: regex pattern to suppress the output when stdout matches


Configuration
-------------

The config file needs `[mute]` section, where `mute` directives are specified.


.. code-block::

    [mute]

    # When a command matched this criteria, the output will be muted.
    # Exit codes and stdout patterns are grouped by "AND", requiring both to match.
    # Multiple sections will be grouped by "OR", so matching any section will suppress the output.
    # stdout is checked by matching with regular expression patterns.

    [[ mute.default ]]
    exit_codes = [0]  # any of exit codes could match

    # OR
    [[ mute.default ]]
    stdout_patterns = [".+ OK .+"]  # stdout matches any listed regex patterns

    # OR
    [[ mute.default ]]
    exit_codes = [1, 2]  # any program that exits with either 1,2 AND prints OK
    stdout_patterns = ["OK"]

    # Command specific settings, overriding default settings, not stacking with default.
    # This applies to any command starting with 'user': 'user' and 'useradd' and 'userdel'
    [[ mute.commands.user ]]
    exit_codes = [0]  # any command starting with "user" will match ONLY when exit code is 0

    # Command specific settings can also be grouped with OR by repeating the settings
    [[ mute.commands.user ]]
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
