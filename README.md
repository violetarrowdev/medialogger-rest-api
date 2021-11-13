# MediaLogger REST API

This project's goal is to serve as a first foray into Go development for me, but also serve as a basic skeleton of a server to support a mobile app I eventually intend to work on. MediaLogger (working title!) will function as a way to organize your media to-do list all in one place: movies you intend to watch, books you plan on reading, games you've always wanted to play, and shows you're behind on can all be tagged, split into sub-lists, and ordered in whatever way you want, all with cloud sync capability. Future integration with APIs for sites IMDB, Goodreads, MyAnimeList, and other similar sites will be considered, if workable!

# Current Features

The API currently enables adding, deleting, and editing media under a user, as well as logging in and out as a user. Adding, removing, and editing lists, as well as creating new users, will be added at a later date.

# Using the API

While not fully feature-complete, the API currently provides the following endpoints. All endpoints with a `:name` parameter must provide a session token encoded in base 64 as the password portion of an `Authentication` header in the request, otherwise the endpoint will return a `401 Unauthorized`.

## /login

### POST

All sessions must begin by receiving a session token through logging in. No other endpoints are at present accessible without a session token.

To log in, make a `POST` request with a `"username"` field and a `"password"` field in the body of the request in `x-www-form-urlencoded` format. If your request is successful, the endpoint will return `200 OK` along with a 32-bit token that must then be saved.

## /users/:name

### GET

The endpoint will returned sanitized user information, which excludes the user's password and unique ID, but will return the associated username, email, saved media, and saved lists.

## /users/:name/password

### POST

To change a user's password, a `POST` request must be made with an `"oldPassword"` and a `"newPassword"` field in the body of the request. If the oldPassword matches the user's saved hashed password, then it will be updated with `newPassword`.

## /users/:name/email

### GET

Returns the user's email.

### POST

If a `"newEmail"` field is provided in the body of the request, the user's email will be updated with that email.

## /users/:name/media

### GET

Returns a list of all media the user has added to their account.

### PUT

Adds a new piece of media to the user's account which must have a unique UID. If the UID is not unique, the request will fail. See the section on [MediaItem formatting]() for more information. Returns `201 Created` when successful.

## /users/:name/media/:uid

### POST

Updates a media item that matches the UID field provided in the URL. The body be a valid `MediaItem`. Returns `200 OK` when successful.

### DELETE

Deletes the media item matching the UID provided as a parameter in the URL. Returns `200 OK` when successful.

## /users/:name/logout

### POST

No body required. Logs the user out of their current session and deletes the server-side session token. Returns `200 OK`.

# Data Formatting

## User

```
{
  "username": string, // (unique)
  "email": string,
  "passwordHash": string,
  "savedMedia": MediaItem[],
  "savedLists": MediaList[]
}
```

## MediaItem

```
{
  "uid": int,
  "title": string,
  "releaseDate": string, // (mm/dd/yyyy, mm/yyyy, or yyyy)
  "medium": string,
  "description": string,
  "thumbnail": string, // (link)
  "rating": int, // (out of 100)
  "linkedPlatform": string,
  "completed": bool,
  "notes": string
}
```

## MediaList

```
{
  "name": string, // (unique)
  "description": string,
  "mediaTypes": string[],
  "contents": MediaItem[]
}
```

## ListOrder

```
{
  "uid": int,
  "order": int
}
```
