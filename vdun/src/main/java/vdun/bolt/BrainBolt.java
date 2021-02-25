package vdun.bolt;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import org.apache.commons.jexl3.JexlBuilder;
import org.apache.commons.jexl3.JexlContext;
import org.apache.commons.jexl3.JexlEngine;
import org.apache.commons.jexl3.JexlExpression;
import org.apache.commons.jexl3.MapContext;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import backtype.storm.task.OutputCollector;
import backtype.storm.task.TopologyContext;
import backtype.storm.topology.OutputFieldsDeclarer;
import backtype.storm.topology.base.BaseRichBolt;
import backtype.storm.tuple.Fields;
import backtype.storm.tuple.Tuple;
import backtype.storm.tuple.Values;

public class BrainBolt extends BaseRichBolt {
    private static final Logger LOG = LoggerFactory.getLogger(BrainBolt.class);
    private JexlEngine jexl;
    private List<JexlExpression> exprs;
    private List<String> policyList;
    private OutputCollector outputCollector;

    public void prepare(Map stormConf, TopologyContext context, OutputCollector collector) {
        outputCollector = collector;

        policyList = (List<String>)stormConf.getOrDefault("policy",  new ArrayList<String>());
        LOG.info("vdun.policy: {}", policyList);
        jexl = new JexlBuilder().create();
        exprs = new ArrayList<JexlExpression>();
        for (String policy : policyList) {
            exprs.add(jexl.createExpression(policy));
        }
    }

    public void declareOutputFields(OutputFieldsDeclarer declarer) {
        declarer.declare(new Fields("feature"));
    }

    public void execute(Tuple tuple) {
        Map<String, Object> feature = (Map<String, Object>)tuple.getValue(0);
        JexlContext jc = new MapContext(feature);
        for (int i = 0; i < policyList.size(); i++) {
            JexlExpression e = exprs.get(i);
            Object result = e.evaluate(jc);
            if (result instanceof Boolean && (Boolean)result == true) {
                feature.put("risk", true);
                feature.put("policy_hit", policyList.get(i));
                outputCollector.emit(new Values(feature));
                return;
            }
        }
    }
}