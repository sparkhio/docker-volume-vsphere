---
Title: Running ELK stack with vDVS
---

The ELK stack consists of ElasticSearch (a distributed full text search engine), Logstash (a data processing pipeline that takes input from a variety of sources and outputs them to a stash) and Kibana (a data visualization plugin for ElasticSearch). In this example, we will deploy an ELK stack with storage provisioned by vSphere Docker Volume Service.

In this setup, we have three instances of ElasticSearch - one master and two data nodes. The data nodes store their data via vSphere Docker Volume Service. The logstash container takes input on port 5000 and outputs it to ElasticSearch. Kibana runs on port 5601 and we use it to view the data stored in ElasticSearch.

NB: In order to avoid out of memory exception in ElasticSearch, the default mmap count limit in the OS has to be increased. Make sure its 262144 or higher, or use this command to set it:

sysctl –w vm.max_map_count=262144

We will also configure logstash to output its data to ElasticSearch. Create these two files in your working directory.

#1 ./logstash/config/logstash.yml

    ---
    ## Default Logstash configuration from logstash-docker.
    ## from https://github.com/elastic/logstash-docker/blob/master/build/logstash/config/logstash.yml
    #
    http.host: "0.0.0.0"
    path.config: /usr/share/logstash/pipeline

    ## Disable X-Pack
    ## see https://www.elastic.co/guide/en/x-pack/current/xpack-settings.html
    ##     https://www.elastic.co/guide/en/x-pack/current/installing-xpack.html#xpack-enabling
    #
    xpack.monitoring.enabled: false

#2 ./logstash/pipeline/logstash.conf

    input {
    	tcp {
    		port => 5000
    	}
    }

    ## Add your filters / logstash plugins configuration here

    output {
    	elasticsearch {
    		hosts => "esmaster:9200"
    		user => elastic
    		password => changeme
    	}
    }

Now, let's create the docker-compose.yml

    version: '2'

    services:

      esmaster:
        image: docker.elastic.co/elasticsearch/elasticsearch:5.5.0 
        container_name: esmaster
        environment:
          ES_JAVA_OPTS: "-Xmx256m -Xms256m"
          cluster.name: "docker-cluster"
          node.master: "true"
          node.data: "false"
        ports:
          - "9200:9200"
          - "9300:9300"
        networks:
          - elk

      es1:
        image: docker.elastic.co/elasticsearch/elasticsearch:5.5.0 
        container_name: es1
        environment:
          ES_JAVA_OPTS: "-Xmx256m -Xms256m"
          cluster.name: "docker-cluster"
          discovery.zen.ping.unicast.hosts: "esmaster"
        volumes:
          - es1data:/usr/share/elasticsearch/data
        links:
          - esmaster
        networks:
          - elk

      es2:
        image: docker.elastic.co/elasticsearch/elasticsearch:5.5.0 
        container_name: es2
        environment:
          ES_JAVA_OPTS: "-Xmx256m -Xms256m"
          cluster.name: "docker-cluster"
          discovery.zen.ping.unicast.hosts: "esmaster"
        volumes:
          - es2data:/usr/share/elasticsearch/data
        links:
          - esmaster
        networks:
          - elk

      logstash:
        image: docker.elastic.co/logstash/logstash:5.5.0
        container_name: logstash
        environment:
          LS_JAVA_OPTS: "-Xmx256m -Xms256m"
        volumes:
          - ./logstash/config/logstash.yml:/usr/share/logstash/config/logstash.yml
          - ./logstash/pipeline:/usr/share/logstash/pipeline
        ports:
          - "5000:5000"
        networks:
          - elk
        depends_on:
          - esmaster

      kibana:
        image: docker.elastic.co/kibana/kibana:5.5.0
        container_name: kibana
        environment:
          - ELASTICSEARCH_URL=http://esmaster:9200
        ports:
          - "5601:5601"
        networks:
          - elk
        depends_on:
          - esmaster

    networks:
      elk:

    volumes:
      es1data:
        driver: vsphere
        driver_opts:
          size: 1Gb  
      es2data:
        driver: vsphere
        driver_opts:
          size: 1Gb

Start the containers now -

docker-compose up

We should be able to see that there are two volumes es1data and es2data which use the vsphere driver.

    root@vm1:~# docker volume ls 
    DRIVER              VOLUME NAME
    vsphere:latest      elasticsearch_es1data@datastore1
    vsphere:latest      elasticsearch_es2data@datastore1

There should be five containers viz esmaster, es1, es2, logstash and kibana running on the server. Verify this with  docker ps -a

Now lets send some test data to logstash.

    root@vm1:~# echo "Test log" | logger -t mylogs --server your-server-ip --tcp --port 5000

Login to kibana at http://your-server-ip:5601. (You may need to login with the default user elastic:changeme). Create the default index in ElasticSearch when prompted to do so and browse to the Discover tab. You should be able to view the entry for "Test log", that we sent to elasticsearch earlier.

To verify persistence, lets kill both the data nodes

    root@vm1:~# docker kill es1 es2

Kibana should now show Internal Server Error when refreshed, as the data node containers are no longer present.

Lets now restart the containers

    root@vm1:~# docker start es1 es2

Now, Kibana should show the test log that we entered earlier.

In this way, we can set up an ElasticSearch cluster with Docker using vDVS as a storage medium.
