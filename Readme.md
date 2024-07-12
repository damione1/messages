# Messages

**Messages** is a content management system that facilitates the creation and distribution of messages to various websites. The system leverages a stack comprising **Golang**, **Templ**, **HTMX**, and **SQLite**, making it efficient and lightweight. Its' a project based on superkit framework (https://github.com/anthdm/superkit).

## Overview

The main objective of the Messages CMS is to centrally manage and dispatch informational messages to multiple websites efficiently. This is particularly useful for scenarios where there is a need to broadcast the same information across different platforms regardless of their underlying technology.

### Key Features:

1. **Central Message Management:**

   - Create, update, and delete messages from a single interface.
   - Each message contains a title, content, language, category (warning, danger, info), and a date range within which it is active.
   - Markdown support for message content formatting.

2. **Multi-language Support:**

   - Currently supports messages in English ("en") and French ("fr").

3. **Message Categories:**

   - Categorize messages as **warning**, **danger**, or **info** to signify their importance.

4. **Time-bound Messages:**

   - Specify a valid date range for each message, after which it will no longer be displayed.

5. **Website Association:**
   - Associate messages with multiple websites.
   - Use the `Origin` header to determine the requesting website and fetch the respective messages in the correct language.
   - Staging mode: A website can be set to staging mode. in this mode he will also receive the messages that are set in the future.

### API Endpoint

The CMS provides an API endpoint to fetch messages dynamically.

#### Endpoint: `/api/messages`

- **Headers:**

  - `Origin`: The origin domain of the requesting website.
  - `Accept-Language`: The language code (e.g., `en` for English, `fr` for French).
  - `Timezone`: Override the timezone of the user (e.g., `Europe/Paris`).

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

1. **Creating and Managing Messages:**

   - Administrators use the CMS interface to create, update, and manage messages. Each message is associated with one or multiple websites and is assigned a language and category.

2. **Message Storage:**

   - Messages are stored in an SQLite database, with relationships to websites they are associated with.

3. **Fetching Messages (Client-side):**

   - Websites make a GET request to the `/api/messages/{language}` endpoint with the appropriate `Origin` header.
   - The CMS verifies the `Origin`, retrieves the list of associated messages, and returns them in the response formatted as JSON.

4. **Displaying Messages on the Website:**
   - The client-side script parses the JSON response and renders the messages on the website accordingly.
   - Messages are displayed in the correct language and formatted using Markdown.

### Example Workflow

1. **Message Creation:**

   - Admin creates a message titled "Maintenance Notice" with content "The site will be under maintenance on Sunday", in English, categorized as a warning, and sets the display date range.

2. **API Request:**

   - Website `http://example.com` makes a request to `/api/messages/en` with `Origin: example.com`.
   - CMS verifies the origin, fetches the relevant messages, and returns them.

3. **Message Display:**
   - The website parses the response and displays the "Maintenance Notice" message to its users.

By centralizing message management and using a robust API, the Messages CMS ensures consistency and efficiency in information dissemination across multiple platforms. This makes it an invaluable tool for administrators aiming to maintain uniform communication across their web properties.

### Client-side Implementation example

#### HTML

```html
<div
  id="message-container"
  class="fixed top-0 inset-x-0 flex flex-col items-center space-y-4 pt-4"
></div>
```

#### JavaScript

JavaScript implementation example using Tailwind CSS for styling:

```javascript
document.addEventListener("DOMContentLoaded", function () {
  const messageContainer = document.getElementById("message-container");
  const apiUrl = `/api/messages`;

  fetch(apiUrl, {
    headers: {
      Origin: window.location.origin, // Change to the origin domain of the requesting website
      "Accept-Language": "en", // Change to the desired language code
      Timezone: "Europe/Paris", // Override the timezone of the user
    },
  })
    .then((response) => response.json())
    .then((data) => {
      if (data.error) {
        console.error("Error fetching messages:", data.error);
        return;
      }
      displayMessages(data.messages);
    })
    .catch((error) => {
      console.error("Error fetching messages:", error);
    });

  function displayMessages(messages) {
    messages.forEach((message) => {
      const messageElement = document.createElement("div");
      const messageTypeClasses = {
        info: "bg-blue-100 border-blue-400 text-blue-700",
        warning: "bg-yellow-100 border-yellow-400 text-yellow-700",
        danger: "bg-red-100 border-red-400 text-red-700",
      };

      messageElement.className = `message ${
        messageTypeClasses[message.type]
      } border-l-4 p-4 w-11/12 md:w-1/2 lg:w-1/3 mx-auto rounded shadow`;
      messageElement.innerHTML = `
                <div class="flex">
                    <div class="flex-shrink-0">
                        <svg class="h-5 w-5 text-${
                          message.type === "info"
                            ? "blue"
                            : message.type === "warning"
                            ? "yellow"
                            : "red"
                        }-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                d="${
                                  message.type === "info"
                                    ? "M13 16h-1v-4H8m0 0H7.5a2.5 2.5 0 110-5h1.5a2.5 2.5 0 010 5h-1.5M17 16h1m0 0h-1m0 0H8.5a2.5 2.5 0 110-5H17a2.5 2.5 0 010 5z"
                                    : message.type === "warning"
                                    ? "M13 16h-1v-4H8m0 0H7.5a2.5 2.5 0 110-5h1.5a2.5 2.5 0 010 5h-1.5M17 16h1m0 0h-1m0 0H8.5a2.5 2.5 0 110-5H17a2.5 2.5 0 010 5z"
                                    : "M13 16h-1v-4H8m0 0H7.5a2.5 2.5 0 110-5h1.5a2.5 2.5 0 010 5h-1.5M17 16h1m0 0h-1m0 0H8.5a2.5 2.5 0 110-5H17a2.5 2.5 0 010 5z"
                                }" />
                        </svg>
                    </div>
                    <div class="ml-3">
                        <h3 class="text-sm font-medium">${message.title}</h3>
                        <div class="mt-2 text-sm">${message.message}</div>
                    </div>
                    <div class="ml-auto pl-3">
                        <button type="button" class="inline-flex text-${
                          message.type === "info"
                            ? "blue"
                            : message.type === "warning"
                            ? "yellow"
                            : "red"
                        }-500 hover:text-white hover:bg-${
        message.type === "info"
          ? "blue"
          : message.type === "warning"
          ? "yellow"
          : "red"
      }-500 rounded p-1" aria-label="Dismiss" onclick="this.parentElement.parentElement.remove();">
                            <svg class="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                                <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"/>
                            </svg>
                        </button>
                    </div>
                </div>
            `;
      messageContainer.appendChild(messageElement);
    });
  }
});
```

### Explanation

1. **HTML Structure:**

   - We added a reference to Tailwind CSS for design purposes.
   - The `message-container` div is fixed at the top of the page and is set to be flex column to stack multiple messages vertically.

2. **JavaScript Implementation:**
   - `messageElement.className` is set dynamically to include Tailwind CSS classes based on the message type (`info`, `warning`, `danger`).
   - We use Tailwind utility classes for styling and responsiveness.
   - The message also includes an SVG icon and dismiss button, styled according to the message type.
   - The dismiss button allows users to remove a message from the view by clicking on it.
