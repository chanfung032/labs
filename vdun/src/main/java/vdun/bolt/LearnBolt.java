package vdun.bolt;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import backtype.storm.Config;
import backtype.storm.task.OutputCollector;
import backtype.storm.task.TopologyContext;
import backtype.storm.topology.OutputFieldsDeclarer;
import backtype.storm.topology.base.BaseRichBolt;
import backtype.storm.tuple.Fields;
import backtype.storm.tuple.Tuple;
import backtype.storm.tuple.Values;
import backtype.storm.utils.TupleUtils;
import vdun.util.ConfigParser;
import vdun.util.URLNormalizer;

public class LearnBolt extends BaseRichBolt {
    private static final Logger LOG = LoggerFactory.getLogger(LearnBolt.class);
    private OutputCollector outputCollector;
    private List<String> urls;
    private Boolean isLearning;
    private int interval;
    private int elapsedMins;
    private int logCount;

    public void prepare(Map stormConf, TopologyContext context, OutputCollector collector) {
        this.outputCollector = collector;
        this.urls = new ArrayList<String>();
        this.isLearning = true;
        this.elapsedMins = 0;
        ConfigParser parser = new ConfigParser(stormConf);
        this.interval = parser.getLong("learn.interval", 60L).intValue();
        this.logCount = parser.getLong("learn.url_count", 2000L).intValue();
    }

    public Map<String, Object> getComponentConfiguration() {
        Map<String, Object> conf = new HashMap<String, Object>();
        conf.put(Config.TOPOLOGY_TICK_TUPLE_FREQ_SECS, 60);
        return conf;
    }

    public void declareOutputFields(OutputFieldsDeclarer declarer) {
        declarer.declare(new Fields("params"));
    }

    public void execute(Tuple tuple) {
        if (TupleUtils.isTick(tuple)) {
            if (!isLearning) {
                elapsedMins += 1;
                if (elapsedMins >= interval) {
                    isLearning = true;
                }
            }
        } else {
            if (!isLearning) {
                return;
            }

            if (urls.size() < logCount) {
                urls.add(tuple.getString(0));
                return;
            }

            URLNormalizer normalizer = new URLNormalizer();
            normalizer.fit(urls);
            try {
                ObjectMapper mapper = new ObjectMapper();
                String params = mapper.writeValueAsString(normalizer);
                LOG.info(params.toString());
                outputCollector.emit(new Values(params));
            } catch (JsonProcessingException e) {
                LOG.error(e.toString());
            }

            elapsedMins = 0;
            isLearning = false;
            urls.clear();
        }
    }
}