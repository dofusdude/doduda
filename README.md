# doduda

Loading and parsing Dofus data.

This project is work in progress - but already usable for some basic tasks.

## Quick Start

*Note: This takes a lot of Docker ressources. If it crashes, increase your memory.*

```bash
docker pull stelzo/doduda:latest

# Download latest Dofus2 data to the ./data directory.
docker run --rm -it -v $PWD/data:/data stelzo/doduda:latest

# Make the data usable for humans.
docker run --rm -it -v $PWD/data:/data stelzo/doduda:latest parse
```

This will produce
- `data/MAPPED_ITEMS.json`
- `data/MAPPED_MOUNTS.json`
- `data/MAPPED_SETS.json`
- `data/MAPPED_RECIPES.json`

- `data/img/item` All item icons
- `data/img/mount` All mount icons
- `data/vector/item` Vector images of all items
- `data/vector/mount` Vector images of all mounts

... and more. Look in your `./data` folder.

