package vdun.bolt;

import java.util.HashMap;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import backtype.storm.Config;
import backtype.storm.task.OutputCollector;
import backtype.storm.task.TopologyContext;
import backtype.storm.topology.OutputFieldsDeclarer;
import backtype.storm.topology.base.BaseRichBolt;
import backtype.storm.tuple.Fields;
import backtype.storm.tuple.Tuple;
import backtype.storm.tuple.Values;
import backtype.storm.utils.TupleUtils;

public class PvBolt extends BaseRichBolt {
    private static final Logger LOG = LoggerFactory.getLogger(PvBolt.class);
    private OutputCollector outputCollector;
    private Map<String, Integer> counter;

    public void prepare(Map stormConf, TopologyContext context, OutputCollector collector) {
        this.outputCollector = collector;
        this.counter = new HashMap<String, Integer>();
    }

    public void declareOutputFields(OutputFieldsDeclarer declarer) {
        declarer.declare(new Fields("domain", "pv"));
    }

    public Map<String, Object> getComponentConfiguration() {
        Map<String, Object> conf = new HashMap<String,Object>();
        conf.put(Config.TOPOLOGY_TICK_TUPLE_FREQ_SECS, 1);
        return conf;
    }

    public void execute(Tuple tuple) {
        if (TupleUtils.isTick(tuple)) {
            for (Map.Entry<String, Integer> e : counter.entrySet()) {
                LOG.debug("pv: {}, {}", e.getKey(), e.getValue());
                outputCollector.emit(new Values(e.getKey(), e.getValue()));
            }
            counter.clear();
        } else {
            String domain = tuple.getString(0);
            counter.put(domain, counter.getOrDefault(domain, 0) + 1);
        }
    }
}