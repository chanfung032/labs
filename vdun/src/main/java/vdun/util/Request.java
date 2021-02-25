package vdun.util;

import java.util.HashMap;
import java.util.Map;

public class Request extends HashMap<String, String> {
   public Request(Map<String, String> jsonMsg) {
       put("ip", jsonMsg.getOrDefault("remote_addr", "-"));
       put("domain", jsonMsg.getOrDefault("http_host", "-"));
       put("upstream_status", jsonMsg.getOrDefault("upstream_status", "-"));
       put("method", jsonMsg.getOrDefault("method", "-"));
       put("status", jsonMsg.getOrDefault("status", "200"));
       put("path", jsonMsg.getOrDefault("request_uri", "-").split("\\?", 2)[0]);
       put("user_agent", jsonMsg.getOrDefault("http_user_agent", "-"));
       put("referer", jsonMsg.getOrDefault("http_referer", "-"));
       put("request_length", jsonMsg.getOrDefault("request_length", "0"));
       put("request_uri", jsonMsg.getOrDefault("request_uri", "-"));
   }
}