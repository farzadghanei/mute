****
Mute
****

.. image:: https://github.com/farzadghanei/mute/workflows/tests/badge.svg
    :target: https://github.com/farzadghanei/mute/actions


``mute`` runs other programs and mutes the output under configured
conditions. A good use case is to keep cron jobs silenced and avoid receiving
emails for known conditions.

It's written in Go, has a small resource overhead with no runtime dependencies.


Usage
-----

.. code-block::

    # by default exit code "0" is muted
    mute bash -c "echo 'this is muted'"
    mute bash -c "echo 'this is printed, exiting with 12'; exit 12"

    # configure mute with environment variables
    env MUTE_EXIT_CODES="4,5" mute bash -c "echo 'muted'; exit 4"
    env MUTE_STDOUT_PATTERN=".*OK.*" mute bash -c "echo 'warning but OK so muted'; exit 1"

``mute`` accepts a command with optional arguments to run. ``mute`` itself
has no arguments but can be configured with a file (in `TOML <https://github.com/toml-lang/toml>`_),
and environment variables. The configuration is validated before running the command.

The exit code of ``mute`` is the exit code of the command it runs.
However ``mute`` exits with 127 (``mute.ExitErrExec``) when failed to execute the commnad,
and with 126 (``mute.ExitErrConf``) when configuration is invalid.


Configuration
-------------

``mute`` can be configured with environment variables, or with a configuration file.
If the environment variables are set, they define the behavior and
the config file is not even checked. If no variables are defined or they are all empty,
then the configuration file is used.

If the configuration file does not exist, or is not accessible (permissions, etc.)
``mute`` continues with the default configuration.

Any accessible configuration should be valid, otherwise ``mute`` exits with ``mute.ExitErrConf``
(also applies to environment variables).


Default Config
==============
When there is no config specified, ``mute`` suppresses output from successful runs, matching
exit code 0 and any output pattern.


Environment Variables
=====================


* ``MUTE_EXIT_CODES``: comma separated list of exit codes to mute (same as ``exit_codes`` in ``mute.default`` config)
* ``MUTE_STDOUT_PATTERN``: regex pattern to suppress the output when stdout matches
* ``MUTE_CONFIG``: absolute/relative path to the config file. default is ``/etc/mute.toml``, no file no issue. an empty value means no config file lookup.


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
-------

``mute`` is an open source project released under the terms of the `MIT license <https://opensource.org/licenses/MIT>`_.
See LICENSE file for more details.
