from flask import render_template, redirect
from flask_appbuilder.models.sqla.interface import SQLAInterface
from flask_appbuilder import ModelView, ModelRestApi
from flask_appbuilder.actions import action

from . import appbuilder, db, models

class CounterApi(ModelRestApi):
    datamodel = SQLAInterface(models.Counter)

appbuilder.add_api(CounterApi)

class FileStorageModelView(ModelView):
    datamodel = SQLAInterface(models.FileStorage)

    label_columns = {
        "name": "Nome do arquivo",
        "file": "O arquivo"
    }

    list_columns = ["name", "file"]


appbuilder.add_view(
    FileStorageModelView,
    "Lista arquivos",
    icon="fa-paperclip",
    category="Exemplos"
)


class CounterModelView(ModelView):
    datamodel = SQLAInterface(models.Counter)

    label_columns = {
        'name': "Nome do contador",
        'value': "Contagem"
    }
    list_columns = ["name", "value"]

    show_fieldsets = [
        ('Sumário', {'fields': ['name', 'value']})
    ]

    @action("increment", "Incrementar contador", "Dale?", "fa-plus")
    def increment(self, items):
        for item in items:
            item.value += 1
            db.session.add(item)
        db.session.commit()
        return redirect(self.get_redirect())

    @action("decrement", "Decrementar contador", "Dale?", "fa-minus")
    def decrement(self, items):
        for item in items:
            item.value -= 1
            db.session.add(item)
        db.session.commit()
        return redirect(self.get_redirect())

appbuilder.add_view(
    CounterModelView,
    "Lista contadores",
    icon="fa-list",
    category="Exemplos"
)


"""
    Create your Model based REST API::

    class MyModelApi(ModelRestApi):
        datamodel = SQLAInterface(MyModel)

    appbuilder.add_api(MyModelApi)


    Create your Views::


    class MyModelView(ModelView):
        datamodel = SQLAInterface(MyModel)


    Next, register your Views::


    appbuilder.add_view(
        MyModelView,
        "My View",
        icon="fa-folder-open-o",
        category="My Category",
        category_icon='fa-envelope'
    )
"""

"""
    Application wide 404 error handler
"""


@appbuilder.app.errorhandler(404)
def page_not_found(e):
    return (
        render_template(
            "404.html", base_template=appbuilder.base_template, appbuilder=appbuilder
        ),
        404,
    )


db.create_all()
