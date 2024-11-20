from flask import render_template, redirect, request, make_response
from flask_appbuilder.models.sqla.interface import SQLAInterface
from flask_appbuilder import ModelView, ModelRestApi
from flask_appbuilder.forms import DynamicForm
from flask_appbuilder.actions import action
from flask_appbuilder.urltools import get_filter_args
from flask_appbuilder.baseviews import expose, expose_api
from flask_appbuilder.security.decorators import has_access_api, has_access, permission_name
from flask_appbuilder.widgets import FormVerticalWidget
from wtforms import Form
from wtforms.fields import FileField

from . import appbuilder, db, models

class CounterApi(ModelRestApi):
    datamodel = SQLAInterface(models.Counter)

appbuilder.add_api(CounterApi)

class FileForm(DynamicForm):
    file = FileField("Arquivo")

class VersionedFileStorage(ModelView):
    datamodel = SQLAInterface(models.VersionedFileStorage)

    label_columns = {
        'name': "Nome do arquivo",
        'version': 'Versão do arquivo',
        'blob': 'Arquivo'
    }

    add_columns = ['file']
    list_columns = ['name', 'version', 'download']
    get_columns = ['name', 'version', 'download']

    @expose("/download/<pk>")
    @has_access
    def download(self, pk):
        item = self.datamodel.get(pk)
        response = make_response(item.blob.data, 200)
        response.headers['Content-Disposition'] = f"attachment; filename={item.name}"
        return response

    def _add(self): # https://github.com/dpgaspar/Flask-AppBuilder/blob/fab9013003a41c4e80da04f072201a8c7cc99187/flask_appbuilder/baseviews.py#L1208
        is_valid_form = True
        get_filter_args(self._filters, disallow_if_not_in_search=False)
        exclude_cols = self._filters.get_relation_cols()
        form = self.add_form.refresh()
        if request.method == 'POST':
            self._fill_form_exclude_cols(exclude_cols, form)
            if form.validate():
                self.process_form(form, True)
                item = self.datamodel.obj()

                try:
                    # print(form.file)
                    file = form._fields['file']

                    # print(dir(file))
                    # print(file.short_name)
                    data = file.raw_data[0]
                    # print('data', data)
                    # print('data', dir(data))
                    # print('data', data.filename)
                    # print('data', data.stream, type(data.stream))

                    item.name = data.filename
                    item.blob = models.DataBlob(
                        data=data.stream.read()
                    )

                    self.pre_add(item)
                except Exception as e:
                    flash(str(e), "danger")
                    print(e)
                else:
                    if self.datamodel.add(item):
                        pass
                        self.post_add(item)
                    flash(*self.datamodel.message)
                finally:
                    return None
            else:
                is_valid_form = False
        if is_valid_form:
            self.update_redirect()
        return self._get_add_widget(form=form, exclude_cols=exclude_cols)

    add_form = FileForm

appbuilder.add_view(
    VersionedFileStorage,
    "Lista arquivos e versões",
    icon='fa-paperclip',
    category="Exemplos"
)

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
