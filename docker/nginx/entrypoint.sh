#!/bin/ash

envsubst '${AP_MQTT_ADAPTER_MQTT_PORT}' < /etc/nginx/snippets/mqtt-upstream-single.conf > /etc/nginx/snippets/mqtt-upstream.conf
envsubst '${AP_MQTT_ADAPTER_WS_PORT}' < /etc/nginx/snippets/mqtt-ws-upstream-single.conf > /etc/nginx/snippets/mqtt-ws-upstream.conf

envsubst '
    ${AP_USERS_HTTP_PORT}
    ${AP_THINGS_HTTP_PORT}
    ${AP_HTTP_ADAPTER_PORT}
    ${AP_WS_ADAPTER_PORT}' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf

exec nginx -g "daemon off;"
