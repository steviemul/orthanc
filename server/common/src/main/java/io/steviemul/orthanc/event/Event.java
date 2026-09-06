package io.steviemul.orthanc.event;

import lombok.Builder;
import lombok.Getter;

import java.time.Instant;
import java.util.List;

@Builder
@Getter
public class Event {

  private Instant timestamp;
  private String eventType;
  private int pid;
  private String process;
  private String path;
  private String source;
  private String cwd;
  private List<Evidence> evidence;
}
