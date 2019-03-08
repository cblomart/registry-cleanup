# registry-clean

Registry Clean plugin cleans a repository in a docker registry.

## defautls

The registry repository name is mapped to ```DRONE_REPO```.

The docker username is mapped to ```DRONE_REPO_OWNER```.

The plugin will match any hexadecimal tag. That is to say that the regex used is ```[0-9A-Fa-f]+``` also refered to as commit tag.

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