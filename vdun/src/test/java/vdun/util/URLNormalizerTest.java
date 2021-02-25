package vdun.util;

import org.junit.Test;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;

import static org.junit.Assert.*;

import java.util.Arrays;

public class URLNormalizerTest {
    @Test
    public void testNormalizeWithoutAnyFit() {
        URLNormalizer normalizer = new URLNormalizer();
        assertEquals("/*/*/*?x=*&y=*&z=*", normalizer.transform("/a/b/c?y=123&x=kkkkkkk&z="));
    }

    @Test
    public void testURLNormalize() {
        String[] urls = new String[]{
            "/gen?recipe=Zhong_Guo_Dian_Ying_Zi_Liao_Guan",
            "/gen?recipe=Aliyun_Changelog",
            "/a/aHR0cDovL29xdjhzOTQxYS5ia3QuY2xvdWRkbi5jb20vVUN0QUlQakFCaVFEM3FqbEVsMVQ1VnBBL0ZUdC13TEFKWUJZLm1wMw==.html",
            "/gen?recipe=Cun_Jin_Bao_Gold_Price",
            "/a/aHR0cDovL29xdjhzOTQxYS5ia3QuY2xvdWRkbi5jb20vVUNFTkd1S3FQYjdXb2pLdWlzaGIxT2tRL2tOUEJCSGZ2SllFLm1wMw==.html",
            "/a/aHR0cDovL29xdjhzOTQxYS5ia3QuY2xvdWRkbi5jb20vVUNFTkd1S3FQYjdXb2pLdWlzaGIxT2tRL2tOUEJCSGZ2SllFLm1wMw==.html",
            "/gen?recipe=Zhi_Hu_Zhuan_Lan&name=WebNotes",
            "/gen?recipe=Zhi_Hu_Zhuan_Lan&name=Lalaby",
            "/gen?recipe=Pai_Qiu_Shao_Nian",
            "/gen?recipe=Jin_Ji_De_Ju_Ren",
            "/gen?recipe=Qiang_Qiang_San_Ren_Xing",
        };

        URLNormalizer normalizer = new URLNormalizer();
        normalizer.fit(Arrays.asList(urls));
        System.out.println(normalizer.transform("/gen?recipe=Zhong_Guo_Dian_Ying_Zi_Liao_Guan"));
        assertEquals("/gen?recipe=*", normalizer.transform("/gen?recipe=Zhong_Guo_Dian_Ying_Zi_Liao_Guan"));
        assertEquals("/a/*.html", normalizer.transform("/a/aHR0cDovL29xdjhzOTQxYS5ia3QuY2xvdWRkbi5jb20vVUNFTkd1S3FQYjdXb2pLdWlzaGIxT2tRL2tOUEJCSGZ2SllFLm1wMw==.html"));
        assertEquals("/gen?name=*&recipe=Zhi_Hu_Zhuan_Lan", normalizer.transform("/gen?recipe=Zhi_Hu_Zhuan_Lan&name=WebNotes"));
    }

    @Test
    public void testSerialization() throws Exception {
        URLNormalizer normalizer = new URLNormalizer();
        normalizer.fit(Arrays.asList(new String[]{"/gen?recipe=xxx", "/gen?recipe=xxx"}));
        assertEquals("/gen?recipe=xxx", normalizer.transform("/gen?recipe=xxx"));

        ObjectMapper mapper = new ObjectMapper();
        // mapper.configure(SerializationFeature.FAIL_ON_EMPTY_BEANS, false);
        String json = mapper.writeValueAsString(normalizer);
        URLNormalizer normalizer1 = mapper.readValue(json, URLNormalizer.class);
        assertEquals("/gen?recipe=xxx", normalizer1.transform("/gen?recipe=xxx"));
    }
}