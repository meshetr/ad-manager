# ad-manager

## Endpoints

Ad endpoints:

| method | path           | descriiption                             |
|--------|----------------|------------------------------------------|
| POST   | /api/v1/ad     | add another ad                           |
| PUT    | /api/v1/ad     | post updated ad information about the ad |
| DELETE | /api/v1/ad/:id | delete ad                                |

Photo endpoints:

| method | path              | descriiption      |
|--------|-------------------|-------------------|
| POST   | /api/v1/photo     | add another photo |
| DELETE | /api/v1/photo/:id | delete photo      |

All endpoints return 200 with empty body.