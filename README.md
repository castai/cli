# cast-cli

CAST AI Command Line Interface
```
Usage:
  cast [command]

Available Commands:
  cluster     Operations on clusters
  help        Help about any command
  login       Login to CAST AI

Flags:
      --api-url string   CAST AI Api URL (default "https://api.cast.ai/v1")
      --debug            Enable debug mode to log api calls
  -h, --help             help for cast
```

### Example

Login

```
cast login --token <YOUR_CAST_AI_TOKEN>
```

List clusters
```
cast cluster list

┌──────────────────────────────────────┬─────────────────┬─────────┬───────────┬────────────────────────────┐
│ #                                    │ NAME            │ STATUS  │ CLOUDS    │ REGION                     │
├──────────────────────────────────────┼─────────────────┼─────────┼───────────┼────────────────────────────┤
│ 19fb46b8-6cb3-4d9b-88af-b70050bde6f2 │ azure-1         │ deleted │ aws azure │ Europe Central (Frankfurt) │
├──────────────────────────────────────┼─────────────────┼─────────┼───────────┼────────────────────────────┤
│ 39cd2fbb-bdf6-4233-bd26-b372c759fb69 │ aws-gcp         │ deleted │ aws gcp   │ Europe Central (Frankfurt) │
├──────────────────────────────────────┼─────────────────┼─────────┼───────────┼────────────────────────────┤
│ 82835a71-7598-4d3c-a8a0-9c629c697d6a │ gcp             │ deleted │ gcp       │ Europe Central (Frankfurt) │
├──────────────────────────────────────┼─────────────────┼─────────┼───────────┼────────────────────────────┤
│ 7d411935-1fb8-4eb8-9e98-bbf4a28693f4 │ help            │ deleted │ gcp       │ Europe Central (Frankfurt) │
└──────────────────────────────────────┴─────────────────┴─────────┴───────────┴────────────────────────────┘
```
