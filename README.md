# ad-manager

## Endpoints

Ad endpoints:

| method | path           | description                             |
|--------|----------------|------------------------------------------|
| POST   | /api/v1/ad     | add another ad                           |
| PUT    | /api/v1/ad     | post updated ad information about the ad |
| DELETE | /api/v1/ad/:id | delete ad                                |

Photo endpoints:

| method | path                        | description      |
|--------|-----------------------------|-------------------|
| POST   | /api/v1/ad/:id/photo        | add another photo |
| DELETE | /api/v1/ad/:ad-id/photo/:id | delete photo      |


## Development database:

```bash
# first time:
docker run -d \
	--name pg-meshetr \
	-e POSTGRES_USER=dbuser \
	-e POSTGRES_PASSWORD=verisikret \
	-e POSTGRES_DB=meshetr \
	-p 5432:5432 \
	postgres

# after:
docker start pg-meshetr
```