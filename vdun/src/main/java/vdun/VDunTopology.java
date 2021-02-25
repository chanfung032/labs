package vdun;

import java.io.FileInputStream;
import java.util.Map;
import java.util.UUID;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;

import backtype.storm.Config;
import backtype.storm.LocalCluster;
import backtype.storm.generated.StormTopology;
import backtype.storm.spout.SchemeAsMultiScheme;
import backtype.storm.topology.TopologyBuilder;
import backtype.storm.tuple.Fields;
import storm.kafka.BrokerHosts;
import storm.kafka.KafkaSpout;
import storm.kafka.SpoutConfig;
import storm.kafka.StringScheme;
import storm.kafka.ZkHosts;
import vdun.bolt.BrainBolt;
import vdun.bolt.DetectBolt;
import vdun.bolt.FilterBolt;
import vdun.bolt.LearnBolt;
import vdun.bolt.OutputBolt;

public class VDunTopology {
	public static void main(String[] args) throws Exception {
		if (args.length != 1) {
			String path = VDunTopology.class.getProtectionDomain().getCodeSource().getLocation().getPath();
			System.out.printf("Usage: storm jar %s CONFIG_FILE\n", path);
			return;
		}
		ObjectMapper mapper = new ObjectMapper(new YAMLFactory());
		Map vdunConf = mapper.readValue(new FileInputStream(args[0]), Map.class);
		System.out.println(vdunConf);

		TopologyBuilder builder = new TopologyBuilder();

		Map<String, Object> inputConf = (Map<String, Object>)vdunConf.get("input");
		BrokerHosts hosts = new ZkHosts((String)inputConf.get("zookeeper"));
		String topic = (String)inputConf.get("topic");
		SpoutConfig spoutConfig = new SpoutConfig(hosts, topic, "/" + topic, UUID.randomUUID().toString());
		spoutConfig.scheme = new SchemeAsMultiScheme(new StringScheme());
		spoutConfig.startOffsetTime = kafka.api.OffsetRequest.LatestTime();
		spoutConfig.ignoreZkOffsets = true ;
		builder.setSpout("input", new KafkaSpout(spoutConfig));

		builder.setBolt("filter", new FilterBolt()).shuffleGrouping("input").allGrouping("learn");
		builder.setBolt("learn", new LearnBolt()).globalGrouping("filter", FilterBolt.URLStreamId);
		builder.setBolt("detect", new DetectBolt()).fieldsGrouping("filter", new Fields("domain", "ip"));
		builder.setBolt("brain", new BrainBolt()).shuffleGrouping("detect");
		builder.setBolt("output", new OutputBolt()).shuffleGrouping("brain");

		Config config = new Config();
		config.setDebug(true);
		config.setMaxTaskParallelism(1);
		config.putAll(vdunConf);

		StormTopology topology = builder.createTopology();
		LocalCluster cluster = new LocalCluster();
		cluster.submitTopology("vdun", config, topology);
	}
}
