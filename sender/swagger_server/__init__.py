import connexion

from swagger_server import encoder

app = connexion.App(__name__, specification_dir='./swagger/')
app.app.json_encoder = encoder.JSONEncoder
app.add_api('swagger.yaml', arguments={'title': 'Sender'}, pythonic_params=True)

application = app.app
