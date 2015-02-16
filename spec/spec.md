Spec
==================

This is the specs for the results crawler webservice. It covers the different objects returned by the webservice and the available endpoints and operations. This document covers the version 1.0 of the api.

All requests begin with the /api/v1 prefix to specify the version of the api to use. For now the only version available is version 1.0.

All requests and responses are encoded in the JSON format. No other format is supported for now.

Entities
------------

###CrawlerConfig
The crawler config object represents the configuration of the crawler for a user.

Property name         | Type   | Description
----------------------|--------|----------------
**userId**            | string | The unique identifier for the user.
**status**            | bool   | If the crawler is enabled.
**code**              | string | The UQAM user identifier.
**nip**               | string | The UQAM user NIP.
**notificationEmail** | string | The email for new results notifications.

###Results
The results object represents all the results for a user.

Property name                                | Type   | Description
---------------------------------------------|--------|-----------------------------------------------------------
**userId**                                   | string | The unique identifier for the user.
**lastUpdate**                               | string | If the crawler is enabled.
**classes[]**                                | list   | List of all the classes.
classes[].**id**                             | string | The unique identifier of the class.
classes[].**name**                           | string | The name of the class. Ex.: MAT1600
classes[].**group**                          | string | The group for the class.
classes[].**year**                           | string | The session of the class. It is represented by a string conaining the year and the session    number. The session numbers are 1 for winter, 2 for summer and 3 for fall. Ex.: 20151 for the winter session of 2015.
classes[].**results[]**                      | list   | The list of all results for the class.
classes[].results[].**name**                 | string | The name of the results. Ex.: Exam 1.
classes[].results[].**normal**               | object | Details of the non-ponderated result.
classes[].results[].normal.**result**        | string | The user's grade for the non-ponderated result. It is a string formatted like: 15/20.
classes[].results[].normal.**average**       | string | The average for the non-ponderated result. It is a string formatted like: 15/20.
classes[].results[].normal.**standardDev**   | string | The standard deviation for the non-ponderated result. It is a string formatted like: 15/20.
classes[].results[].**weighted**             | object | Details of the ponderated result.
classes[].results[].weighted.**result**      | string | The user's grade for the ponderated result. It is a string formatted like: 15/20.
classes[].results[].weighted.**average**     | string | The average for the ponderated result. It is a string formatted like: 15/20.
classes[].results[].weighted.**standardDev** | string | The standard deviation for the ponderated result. It is a string formatted like: 15/20


API
------------

###Authentication

####Login

Endpoint: /api/v1/auth/login

Methods: POST

Request body:

Property name         | Type   | Description
----------------------|--------|----------------
**email**             | string | User email.
**password**          | string | User password.
**deviceType**        | int    | The type of device: 0 web, 1 iOS, 2 Android.

Response: 

####Register

Register sends a request to register a user. 

Endpoint: /api/v1/auth/register

Methods: POST

Request body:

Property name           | Type   | Description
------------------------|--------|----------------
**email**               | string | User email.
**password**            | string | User password.
**firstName**           | string | User first name.
**lastName**            | string | User last name.
**notificationToken**   | string | iOS or Android notification token.
**deviceType**          | int    | The type of device: 0 web, 1 iOS, 2 Android.

####Logout

Endpoint: /api/v1/auth/logout

Methods: POST

###Crawler

####Configuration

Endpoint: /api/v1/crawler/config

Methods: GET, POST

####Classes

Endpoint: /api/v1/crawler/class/:classId

Params: classId, the id of the class to edit or delete.

Methods: GET, POST, PUT, DELETE

####Refresh

Endpoint: /api/v1/crawler/refresh
Methods: POST
