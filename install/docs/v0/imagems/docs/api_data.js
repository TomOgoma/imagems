define({ "api": [
  {
    "type": "put",
    "url": "/upload/base64",
    "title": "Upload Base64 Image",
    "name": "NewB64Image",
    "version": "0.1.0",
    "permission": [
      {
        "name": "owner"
      }
    ],
    "group": "Service",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          },
          {
            "group": "Header",
            "optional": false,
            "field": "Authorization",
            "description": "<p>contains Bearer with JWT e.g. &quot;Bearer jwt.val.here&quot;</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "JSON": [
          {
            "group": "JSON",
            "type": "String",
            "optional": false,
            "field": "folder",
            "description": "<p>The folder to place the image in.</p>"
          },
          {
            "group": "JSON",
            "type": "String",
            "optional": false,
            "field": "image",
            "description": "<p>The base64 encoded image string.</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "200": [
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "time",
            "description": "<p>Most recent server time as an ISO8601 string.</p>"
          },
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "URL",
            "description": "<p>The URL to the uploaded image.</p>"
          }
        ]
      }
    },
    "filename": "pkg/handler/http/http.go",
    "groupTitle": "Service"
  },
  {
    "type": "put",
    "url": "/upload",
    "title": "Upload Image",
    "name": "NewImage",
    "version": "0.1.0",
    "permission": [
      {
        "name": "owner"
      }
    ],
    "group": "Service",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          },
          {
            "group": "Header",
            "optional": false,
            "field": "Authorization",
            "description": "<p>contains Bearer with JWT e.g. &quot;Bearer jwt.val.here&quot;</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "Form": [
          {
            "group": "Form",
            "type": "String",
            "optional": false,
            "field": "folder",
            "description": "<p>The folder to place the image in.</p>"
          },
          {
            "group": "Form",
            "type": "File",
            "optional": false,
            "field": "image",
            "description": "<p>file input containing upload image</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "200": [
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "time",
            "description": "<p>Most recent server time as an ISO8601 string.</p>"
          },
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "URL",
            "description": "<p>The URL to the uploaded image.</p>"
          }
        ]
      }
    },
    "filename": "pkg/handler/http/http.go",
    "groupTitle": "Service"
  },
  {
    "type": "get",
    "url": "/status",
    "title": "Status",
    "name": "Status",
    "version": "0.1.0",
    "group": "Service",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "200": [
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "name",
            "description": "<p>Micro-service name.</p>"
          },
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "version",
            "description": "<p>http://semver.org version.</p>"
          },
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "description",
            "description": "<p>Short description of the micro-service.</p>"
          },
          {
            "group": "200",
            "type": "String",
            "optional": false,
            "field": "canonicalName",
            "description": "<p>Canonical name of the micro-service.</p>"
          }
        ]
      }
    },
    "filename": "pkg/handler/http/http.go",
    "groupTitle": "Service"
  },
  {
    "type": "get",
    "url": "/{userID}/{folder}/{imageName}",
    "title": "View Image",
    "name": "ViewImage",
    "version": "0.1.0",
    "permission": [
      {
        "name": "any with API key"
      }
    ],
    "group": "Service",
    "header": {
      "fields": {
        "Header": [
          {
            "group": "Header",
            "optional": false,
            "field": "x-api-key",
            "description": "<p>the api key</p>"
          }
        ]
      }
    },
    "parameter": {
      "fields": {
        "Query": [
          {
            "group": "Query",
            "type": "String",
            "optional": false,
            "field": "userID",
            "description": "<p>The userID of the image owner.</p>"
          },
          {
            "group": "Query",
            "type": "String",
            "optional": false,
            "field": "folder",
            "description": "<p>The folder containing the image.</p>"
          },
          {
            "group": "Query",
            "type": "String",
            "optional": false,
            "field": "imageName",
            "description": "<p>The name of the image.</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "200": [
          {
            "group": "200",
            "type": "ImageFile",
            "optional": false,
            "field": "file",
            "description": "<p>The requested image file or xml listing of files contained in the specified folder.</p>"
          }
        ]
      }
    },
    "filename": "pkg/handler/http/http.go",
    "groupTitle": "Service"
  }
] });
