package vdun.bolt;

import java.util.Map;

import org.apache.commons.lang.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.fasterxml.jackson.databind.ObjectMapper;

import backtype.storm.task.TopologyContext;
import backtype.storm.topology.BasicOutputCollector;
import backtype.storm.topology.OutputFieldsDeclarer;
import backtype.storm.topology.base.BaseBasicBolt;
import backtype.storm.tuple.Fields;
import backtype.storm.tuple.Tuple;
import backtype.storm.tuple.Values;
import vdun.util.Request;
import vdun.util.URLNormalizer;

public class FilterBolt extends BaseBasicBolt {
    public static final String URLStreamId = "urls";
    private static final Logger LOG = LoggerFactory.getLogger(FilterBolt.class);
    private ObjectMapper json;
    private URLNormalizer normalizer;

    public void prepare(Map stormConf, TopologyContext context) {
        json = new ObjectMapper();
        normalizer = new URLNormalizer();
    }

    public void execute(Tuple tuple, BasicOutputCollector collector) {
        String from = tuple.getSourceComponent();
        if (from.equals("learn")) {
            try {
                ObjectMapper mapper = new ObjectMapper();
                normalizer = mapper.readValue(tuple.getString(0), URLNormalizer.class);
            } catch (Exception e) {
                LOG.error(e.toString());
            }
            return;
        }

        // 解析并清理日志（直接用 Nginx 生成的 json 格式日志, 未对 “\” 进行转义）
        String log = tuple.getString(0);
        Request request;
        try {
            Map jsonMsg = json.readValue(StringUtils.replace(log, "\\x", "\\\\x"), Map.class);
            request = new Request(jsonMsg);
        } catch (Exception e) {
            LOG.error("json decode failed", e);
            return;
        }
        
        // TODO: 过滤不需要检测的日志
        String domain = (String)request.get("domain");
        String ip = (String)request.get("ip");
        request.put("request_pattern", normalizer.transform(request.get("request_uri")));

        collector.emit(new Values(domain, ip, request));
        collector.emit(URLStreamId, new Values(request.get("request_uri")));
    }

    public void declareOutputFields(OutputFieldsDeclarer declarer) {
        declarer.declare(new Fields("domain", "ip", "request"));
        declarer.declareStream(URLStreamId, new Fields("url"));
    }
}