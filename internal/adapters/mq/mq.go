package mq

import (
	"context"

	"gitlab.com/g6834/team31/auth/pkg/logging"
	"gitlab.com/g6834/team31/mailsender/internal/config"
	"gitlab.com/g6834/team31/mailsender/internal/domain/models"
	"gitlab.com/g6834/team31/mailsender/internal/ports"
	"gitlab.com/g6834/team31/tasks/pkg/grpc_task"
	"gitlab.com/g6834/team31/tasks/pkg/mq"
	"gitlab.com/g6834/team31/tasks/pkg/mq/types"
	"google.golang.org/protobuf/proto"
)

type MQClient struct {
	logger            *logging.Logger
	consumerMailTopic mq.Consumer
	producerMailTopic mq.Producer
	service           ports.MailSender
}

func NewMQClient(cMail mq.Consumer, pMail mq.Producer, l *logging.Logger) *MQClient {
	return &MQClient{
		consumerMailTopic: cMail,
		producerMailTopic: pMail,
		logger:            l,
	}
}

func BuildMQClietn(cfg config.Kafka, service ports.MailSender, logger *logging.Logger) *MQClient {
	consumerMailTopic, err := mq.NewConsumer([]string{cfg.URL}, cfg.MailTopic, "0")
	if err != nil {
		logger.Fatal().Err(err).Msg("couldn't init mail topic consumer")
	}
	ProducerMailTopic, err := mq.NewProducer([]string{cfg.URL}, cfg.MailTopic)
	if err != nil {
		logger.Fatal().Err(err).Msg("couldn't init task topic consumer")
	}
	return &MQClient{
		consumerMailTopic: consumerMailTopic,
		producerMailTopic: ProducerMailTopic,
		service:           service,
		logger:            logger,
	}
}

func (m *MQClient) RunMQ(ctx context.Context, producerErrChan chan error) chan error {
	errChanOut := make(chan error)
	go m.ConsumeMailTopic(ctx, errChanOut)
	// go m.ProduceMailTopic(ctx, producerErrChan, errChanOut)
	return errChanOut
}

func (m *MQClient) ConsumeMailTopic(ctx context.Context, errChan chan error) {
	for {
		_, err := m.consumerMailTopic.ReadAndCommit(ctx, m.operateMailEvent)
		if err != nil {
			errChan <- err
		}
	}
}

func (m *MQClient) operateMailEvent(ctx context.Context, msg types.Message) error {
	var mailInp grpc_task.Mail
	err := proto.Unmarshal(msg.Value, &mailInp)
	if err != nil {
		m.logger.Warn().Err(err).Msg("operateMailEvent error while unmarshaling kafka event")
		return err
	}
	mail := models.Email{
		Header: mailInp.Header,
		Body:   mailInp.Body,
	}
	m.service.AcceptMail(mail)
	m.logger.Debug().Msgf("mail accepted by service %v", mail)
	return nil
}

func (m *MQClient) ProduceMailTopic(ctx context.Context, sourceErrChan, errChanOut chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-sourceErrChan:
			if err != nil {
				msg := types.Message{
					Key:   []byte("service mailsender: couldn't send email"),
					Value: []byte(err.Error()),
				}
				m.producerMailTopic.SendMessage(ctx, []types.Message{msg}, 0)
			}
		}
	}
}
