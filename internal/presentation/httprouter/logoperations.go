package httprouter

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	schemalog "github.com/n-r-w/log-server/api/schema/schema.log"
	"github.com/n-r-w/log-server/internal/domain/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Добавить в лог
func (router *HTTPRouter) addLogRecord() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &[]model.LogRecord{}
		// парсим входящий json
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			router.respondError(w, r, http.StatusBadRequest, err)

			return
		}

		if err := router.domain.LogUsecase.Insert(req); err != nil {
			router.respondError(w, r, http.StatusForbidden, err)

			return
		}

		router.respond(w, r, http.StatusCreated, nil)
	}
}

// Получить записи из лога
func (router *HTTPRouter) getLogRecords() http.HandlerFunc {
	type requestParams struct {
		TimeFrom time.Time `json:"timeFrom"`
		TimeTo   time.Time `json:"timeTo"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &requestParams{
			TimeFrom: time.Time{},
			TimeTo:   time.Time{},
		}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			router.respondError(w, r, http.StatusBadRequest, err)

			return
		}

		records, err := router.domain.LogUsecase.Find(req.TimeFrom, req.TimeTo)
		if err != nil || records == nil {
			router.respondError(w, r, http.StatusInternalServerError, err)

			return
		}

		if len(*records) == 0 {
			router.respond(w, r, http.StatusOK, nil)

			return
		}

		if r.Header.Get(binaryFormatHeaderName) == binaryFormatHeaderProtobuf {
			w.Header().Add(binaryFormatHeaderName, binaryFormatHeaderProtobuf)
			// клиент хочет Protobuf
			mRecords := &schemalog.LogRecords{
				Records: nil,
			}

			for _, r := range *records {
				mRecord := &schemalog.LogRecord{
					Id:       r.ID,
					LogTime:  timestamppb.New(r.LogTime),
					RealTime: timestamppb.New(r.RealTime),
					Level:    uint32(r.Level),
					Message1: r.Message1,
					Message2: r.Message2,
					Message3: r.Message3,
				}
				mRecords.Records = append(mRecords.Records, mRecord)
			}

			out, err := proto.Marshal(mRecords)
			if err != nil {
				log.Fatalln("protobuf error")
			}

			router.respondBinary(w, r, http.StatusOK, out, binaryFormatLogs, true)

			return
		}

		// отдаем с gzip сжатием если клиент это желает
		router.respondCompressed(w, r, http.StatusOK, records)
	}
}
