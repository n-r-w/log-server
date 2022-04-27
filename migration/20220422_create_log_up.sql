CREATE TABLE public.log (
  id bigserial,
  record_timestamp timestamp without time zone NOT NULL,
  real_timestamp timestamp without time zone NOT NULL DEFAULT now(),
  level integer NOT NULL,
  message1 text NOT NULL,
  message2 text,
  message3 text,
  CONSTRAINT log_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_log_record_timestamp
    ON public.log USING btree
    (record_timestamp ASC NULLS LAST)
;