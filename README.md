# INFO

create a png image for social networks based on a template

## HOW TO

1. Prepare options as a JSON file

```json
{
    "label": "text",
    "font_path": "path",
    "font_size": 12,
    "background_image": "path",
    "label_color": "#000fff",
    "logos": [],
    "save_path": "path"
}
```

The fields are all pretty understandable. The `logos` field is a list of dictionaries

The logos will be all set at the bottom evenly spaced.

2. Run the script

```bash
classicum-img-template options.json
```

options.json is the path to the options file
