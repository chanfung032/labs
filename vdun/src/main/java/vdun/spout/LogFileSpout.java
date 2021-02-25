package vdun.spout;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.Charset;
import java.util.List;
import java.util.Map;

import org.apache.commons.io.IOUtils;

import backtype.storm.spout.SpoutOutputCollector;
import backtype.storm.task.TopologyContext;
import backtype.storm.topology.OutputFieldsDeclarer;
import backtype.storm.topology.base.BaseRichSpout;
import backtype.storm.tuple.Fields;
import backtype.storm.tuple.Values;

public class LogFileSpout extends BaseRichSpout {
    
    private SpoutOutputCollector outputCollector;
    private List<String> logs;

    public void declareOutputFields(OutputFieldsDeclarer declarer) {
        declarer.declare(new Fields("log"));
    }

    public void open(Map conf, TopologyContext context, SpoutOutputCollector collector) {
        outputCollector = collector;

        try {
            File file = new File("/vagrant/nginx.log");
            InputStream inputStream = new FileInputStream(file);
            logs = IOUtils.readLines(inputStream, Charset.defaultCharset().name());
        } catch (IOException e) {
            throw new RuntimeException(e);
        }
    }

    public void nextTuple() {
        for (String log : logs) {
            outputCollector.emit(new Values(log));
            try {
                Thread.sleep(1000);
            } catch (InterruptedException ex) {
                Thread.currentThread().interrupt();
            }
        }
    }
}