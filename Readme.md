# Messages

**Messages** is a content management system that facilitates the creation and distribution of messages to various websites. The system leverages a stack comprising **Golang**, **Templ**, **HTMX**, and **SQLite**, making it efficient and lightweight. It's a project based on the [superkit](https://github.com/anthdm/superkit) framework.

## Installation

The Docker package is already built. You can directly use the following command:

```bash
docker run -d \
  -p 3030:3001 \
  -v messages_db:/app/db \
  --name messages_app \
  -e SUPERKIT_ENV=production \
  -e APP_NAME=Messages \
  -e HTTP_LISTEN_ADDR=:3001 \
  -e DB_DRIVER=sqlite3 \
  -e DB_NAME=db/app.db \
  -e SUPERKIT_SECRET=$(openssl rand -base64 32) \
  ghcr.io/damione1/messages:latest
```

The container will use the folder `messages_db` as storage to host the SQLite database.

Finally go to `localhost:3030`. The first user to register will be admin. There is no confirmation email, so once registered, return to the login page and proceed.

![login](https://github.com/user-attachments/assets/eda99ea4-f5d2-4d67-83a2-c970ba8683b4)

## Overview

The main objective of the Messages CMS is to centrally manage and dispatch informational messages to multiple websites efficiently. This is particularly useful for scenarios where there is a need to broadcast the same information across different platforms regardless of their underlying technology. The client (a website) makes a GET call to this API and receives a JSON array of messages to display. If there’s nothing to display, the array will be empty.

### Admin UI:

- Create, update, and delete messages from a single interface.
- Each message contains a title, content, language, category (warning, danger, info), a date range within which it is active, and the selection of domains to broadcast the message.
- Markdown support for message content formatting.
- UI available in French and English.

![admin](https://github.com/user-attachments/assets/bcc8fdf7-e832-4c03-90da-26693f9a505a)

### API

The central feature is the API endpoint to fetch messages dynamically.

#### Endpoint: `/api/messages`

- **Headers:**

  - `Origin`: The origin domain of the requesting website.
  - `Accept-Language`: The language code (e.g., `en` for English, `fr` for French) based on the language defined in the message.
  - `Timezone`: Override the timezone of the client website (e.g., `Europe/Paris`).

- **Responses:**
  - **Success (200)**
    ```json
    {
      "domain": "example.com",
      "messages": [
        {
          "title": "Important Update",
          "message": "<p>Here is an important update...</p>",
          "type": "info"
        },
        ...
      ]
    }
    ```
  - **Error (400)**
    ```json
    {
      "domain": "example.com",
      "messages": [],
      "error": "Invalid domain"
    }
    ```

### Workflow

1. **Adding one or more websites:**
   - Add one or more websites to broadcast to. If this is a staging website, check the checkbox. This will return the message independently of the selected date, for preview purposes.
2. **Creating and Managing Messages:**
   - Use the CMS interface to create, update, and manage messages. Each message is associated with one or multiple websites and is assigned a language and category.
3. **Fetching Messages (Client-side):**
   - Websites make a GET request to the `/api/messages` endpoint with the appropriate `Origin` header.
   - The CMS verifies the `Origin`, retrieves the list of associated messages, and returns them in the response formatted as JSON.
4. **Displaying Messages on the Website:**
   - The client-side script parses the JSON response and renders the messages on the website accordingly.
   - Messages are displayed in the correct language and formatted using Markdown.

By centralizing message management and using a robust API, the Messages CMS ensures consistency and efficiency in information dissemination across multiple platforms. This makes it an invaluable tool for administrators aiming to maintain uniform communication across their web properties.
