package io.steviemul.orthanc.event;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.time.Instant;
import java.util.List;

@Builder
@NoArgsConstructor
@AllArgsConstructor
@Getter
@Setter
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
