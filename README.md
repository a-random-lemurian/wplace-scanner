# wplace-scanner
Scans an area on WPlace by requesting tiles in a certain area once in a while. For timelapse and history purposes.

## Configuration
Copy `config.yml.example` to `config.yml`. The defaults are sane.

- `tile_server` - The URL of the tile server where Wplace stores the tiles. Use {x} and {y} as placeholders for the actual tile numbers.
- `zoom_level` - The zoom level to use for calculating the X and Y. Wplace uses only zoom 11 for this.
- `output` - The directory where the tiles will be saved. The program will create YYYYMMDD-HHMMSS folders here.
- `user_agent` - The user agent to use when making requests for the tiles.
- `frequency` - How often to download the tiles.
- `bbox` - The coverage area for the scanner. Attempting to scan the entire canvas is a *bad* idea.
- `stitch_tiles` - Whether to generate a stitched version of all downloaded tiles. Be wary of memory use when scanning large areas!

To calculate a bounding box, use the tool at [http://bboxfinder.com/](http://bboxfinder.com/). In the bottom-right of the website, select Lat / Lng as the coordinate format. Select the square drawing tool and copy the coordinates labeled "Box". Then, put them in-between the square brackets in the configuration file. It is also recommended to add the name of your project of sorts after `tiles/` to keep things organized, for example, `port-elizabeth` or `extremely-huge-art-timelapse`.

## Concerns
Please, be respectful of the servers and the developers.

This program only makes requests for the pixelart tiles, not the vector tiles in the background, reducing load on the servers. Furthermore, the fact that this thing only downloads a set amount of tiles every 60 minutes means that the load this tool places is very low. A single human browsing the map generates plenty more load than this tool.

## Copyright
This program is licensed under the MIT License. See [`LICENSE`](./LICENSE) for details.
