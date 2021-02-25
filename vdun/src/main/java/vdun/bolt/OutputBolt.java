package vdun.bolt;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.hubspot.jinjava.Jinjava;

import backtype.storm.Config;
import backtype.storm.task.OutputCollector;
import backtype.storm.task.TopologyContext;
import backtype.storm.topology.OutputFieldsDeclarer;
import backtype.storm.topology.base.BaseRichBolt;
import backtype.storm.tuple.Tuple;
import backtype.storm.utils.TupleUtils;

import okhttp3.Call;
import okhttp3.Callback;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;

public class OutputBolt extends BaseRichBolt {
    private static final Logger LOG = LoggerFactory.getLogger(OutputBolt.class);
    private Map<String, List<String>> buffer;
    private List<Map<String, String>> actions;
    private Jinjava jinjava;
    private OkHttpClient httpClient;

    public void prepare(Map stormConf, TopologyContext context, OutputCollector collector) {
        buffer = new HashMap<String,List<String>>();
        actions = (List<Map<String, String>>)stormConf.get("action");
        if (actions == null) {
            actions = new ArrayList<Map<String,String>>();
        }
        jinjava = new Jinjava();
        httpClient = new OkHttpClient();
    }

    public void declareOutputFields(OutputFieldsDeclarer declarer) {
    }

    public Map<String, Object> getComponentConfiguration() {
        Map<String, Object> conf = new HashMap<String, Object>();
        conf.put(Config.TOPOLOGY_TICK_TUPLE_FREQ_SECS, 1);
        return conf;
    }

    public void execute(Tuple tuple) {
        if (TupleUtils.isTick(tuple)) {
            executeActions(buffer);
            buffer.clear();
        } else {
            Map<String, Object> feature = (Map<String, Object>)tuple.getValue(0);
            String domain = (String)feature.get("domain");
            List<String> ips = buffer.get(domain);
            if (ips == null) {
                ips = new ArrayList<String>();
                buffer.put(domain, ips);
            }
            ips.add((String)feature.get("ip"));
        }
    }

    private void executeActions(Map<String, List<String>> buffer) {
        for (Entry<String, List<String>> b : buffer.entrySet()) {
            Map<String, Object> context = new HashMap<String,Object>();
            context.put("domain", b.getKey());
            context.put("ips", b.getValue());

            for (Map<String, String> action : actions) {
                String url = action.get("url");
                String data = action.get("data");
                String method = action.getOrDefault("method", "GET");
                if (url == null) {
                    LOG.error("invalid action, missing url: '{}'", action);
                    continue;
                }

                url = jinjava.render(url, context);
                if (data != null) {
                    data = jinjava.render(data, context);
                }
                LOG.info("url: {}, data: {}", url, data);

                Request.Builder builder = new Request.Builder();
                builder.url(url);
                RequestBody body = data != null ? RequestBody.create(null, data) : null;
                builder.method(method, body);
                httpClient.newCall(builder.build()).enqueue(new Callback() {
                    public void onFailure(Call call, IOException e) {
                        e.printStackTrace();
                    }
                    public void onResponse(Call call, Response response) throws IOException {
                        LOG.info("action response: {}", response);
                    }
                });
            }
        }
    }
}