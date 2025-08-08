from argparse import ArgumentParser
import mlflow
import random
import tempfile
from pathlib import Path

parser = ArgumentParser()
parser.add_argument('--mlflow-url', type=str)
parser.add_argument('--mlflow-experiment', type=str, default="el_modelo_malaco")

subparsers = parser.add_subparsers(required=True, title="Subcomandos")

class TestModel(mlflow.pyfunc.PythonModel):

    def load_context(self, context: mlflow.pyfunc.model.PythonModelContext):
        self.model = Path(context.artifacts['model']).read_text()

    def predict(self, context, model_input):
        return self.model

def train(args):
    with mlflow.start_run() as run:
        for i in range(10):
            value = random.random() * i
            mlflow.log_metric('num', value, step=i)
            f = Path(tempfile.mktemp())
            f.mkdir(parents=True, exist_ok=True)
            f = f / "item.txt"
            f.write_text(str(value))
            mlflow.pyfunc.log_model(
                name="teste",
                python_model=TestModel(),
                artifacts={ "model": str(f) },
                step=i,
                model_config={
                    "teste": ['eoq', 'trabson', 'nhaa']
                }
            )
            f.unlink()
    print('train', args)

train_parser = subparsers.add_parser('train')
train_parser.set_defaults(fn=train)
train_parser.description = "Treino"

def inference(args):
    model = mlflow.pyfunc.load_model(model_uri=args.mlflow_model_uri)
    res = model.predict([])
    print('res', res)
inference_parser = subparsers.add_parser('inference')
inference_parser.add_argument('--mlflow-model-uri', required=True)
inference_parser.set_defaults(fn=inference)
inference_parser.description = "Inferência"



def main():
    args = parser.parse_args()
    if args.mlflow_url is not None:
        mlflow.set_tracking_uri(args.mlflow_url)
    experiment = mlflow.get_experiment_by_name(args.mlflow_experiment)
    if experiment is None:
        experiment = mlflow.create_experiment(args.mlflow_experiment)
    mlflow.set_experiment(args.mlflow_experiment)
    args.fn(args)

if __name__ == "__main__":
    main()
