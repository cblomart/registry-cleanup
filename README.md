# registry-clean

[![Build Status](https://cloud.drone.io/api/badges/cblomart/registry-cleanup/status.svg)](https://cloud.drone.io/cblomart/registry-cleanup)

Registry Clean plugin cleans a repository in a docker registry.

## CLI

Bellow the help when used on commandline

```
$ registry-cleanup --help
NAME:
   registry-cleanup - Clean a registry repository from lingering tags/images

USAGE:
   registry-cleanup.exe [global options] command [command options] [arguments...]

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --username value, -u value  Docker username [$PLUGIN_USERNAME, $DRONE_REPO_OWNER]
   --password value, -p value  Docker password [$PLUGIN_PASSWORD]
   --repo value, -r value      Repository to target [$PLUGIN_REPO, $DRONE_REPO]
   --registry value            Registry to target (default: "https://cloud.docker.com") [$PLUGIN_REGISTRY]
   --regex value               Clean Tags that match regex (default: "^[0-9A-Fa-f]+$") [$PLUGIN_REGEX]
   --min value, -m value       Minimum number of tags/images to keep (default: 3) [$PLUGIN_MIN]
   --max value, -M value       Maximum age of tags/images (default: 360h0m0s) [$PLUGIN_MAX]
   --verbose                   Show verbose information [$PLUGIN_VERBOSE]
   --dryrun                    Dry run [$PLUGIN_DRYRUN]
   --dump                      Dump network requests [$PLUGIN_DUMP]
   --help, -h                  show help
   --version, -v               print the version
```

## defautls

The registry repository name is mapped to ```DRONE_REPO```.

The docker username is mapped to ```DRONE_REPO_OWNER```.

The plugin will match any hexadecimal tag. That is to say that the regex used is ```^[0-9A-Fa-f]+$``` also refered to as commit tag.

The plugin will keep at least 3 images matching the regex.

The plugin will delete images matching the regex older than 15 days.

## examples

The following pipeline configuration will use the defaults:

```yaml
kind: pipeline
name: default

steps:
- name: registry-clean
  image: cblomart/registry-clean
  settings:
    password: XXXXXX
```

The following example will target a non default repositrory:

```yaml
kind: pipeline
name: default

steps:
- name: registry-clean
  image: cblomart/registry-clean
  settings:
    password: pirate
    repo: foo/bar
```

The following example will target a non default repository with a non default username:

```yaml
kind: pipeline
name: default

steps:
- name: registry-clean
  image: cblomart/registry-clean
  settings:
    username: john
    password: XXXXXXXXXX
    repo: foo/bar
```

The following example uses custom registry:

>
> Custom registry needs to have [delete enabled](https://docs.docker.com/registry/configuration/#delete)
>

```yaml
kind: pipeline
name: default

steps:
- name: registry-clean
  image: cblomart/registry-clean
  settings:
    username: lazy
    password: pirate
    registry: http//registry.mycompany.com:9000
    repo: foo/bar
```

The following example will keep a minimum of 5 images and delete images older than 7 days

```yaml
kind: pipeline
name: default

steps:
- name: registry-clean
  image: cblomart/registry-clean
  settings:
    password: XXXXXX
    min: 5
    max: 7d
```

#  License

Copyright (c) 2019 cblomart

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.