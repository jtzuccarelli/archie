BEGIN;

UPDATE calls
   SET status = 'complete'
 WHERE status = 'pending'
   AND transcript IS NOT NULL;

COMMIT;
