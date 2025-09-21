use telegram_bot::*;
use futures::{StreamExt, future::lazy};

#[derive(Debug)]
enum BotError {
    NoTokenError,
    TelegramLibError(telegram_bot::Error)
}

impl From<telegram_bot::Error> for BotError {
    fn from(e: telegram_bot::Error) -> BotError {
        BotError::TelegramLibError(e)
    }
}

fn main() {
    use tokio_compat::runtime::current_thread::Runtime;
    let mut runtime = Runtime::new().unwrap();
    println!("Iniciado");
    runtime.block_on_std(async {
        run().await.unwrap();
        ()
    });
    println!("Finalizado");
}

async fn run() -> Result<(), BotError> {
    println!("Rodando");
    let token = match std::env::var("TELEGRAM_BOT") {
        Ok(token) => Ok(token),
        Err(_) => Err(BotError::NoTokenError)
    }?;
    let bot = Api::new(token);
    let mut stream = bot.stream();
    while let Some(update) = &stream.next().await {
        let update = update.as_ref().unwrap();
        match &update.kind {
            UpdateKind::Message(message) => {
                match message.kind {
                    MessageKind::Text {ref data, .. } => {
                        bot.send(message.text_reply(format!("Hello, {}", &message.from.first_name))).await?;
                        println!("{:?} sent {:?}", &message.from.username, data);
                    },
                    _ => {}
                }
            },
            _ => {}
        }
    }
    Ok(())
}
