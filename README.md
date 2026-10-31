## Containers

_A personal opiniated collection of container images_

Translator images are built from private Forgejo release tags. The Renovate
workflow checks this repository every six hours and supports manual dispatch;
merged version updates trigger image publication. See the
[Translator build and runtime instructions](apps/translator/README.md).

[TV guides](apps/tv-guides/README.md) follows the same tagged-source build flow
with an unencrypted image and a persistent SQLite working directory.

## Credits

_Completely and 100% inspired by [Home Operations - Containers](https://github.com/home-operations/containers)_
