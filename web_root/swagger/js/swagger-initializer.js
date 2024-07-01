window.onload = function() {
  //<editor-fold desc="Changeable Configuration Block">

  // the following lines will be replaced by docker/configurator, when it runs in a docker-container
  window.ui = SwaggerUIBundle({
    url: "/swagger/js/swagger.json",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl,
      function() {
        return {
          components: {
            StringParam: function(props) {
              return React.createElement('textarea', {
                value: props.value,
                onChange: e => props.onChange(e.target.value),
                rows: 4,
                cols: 50
              });
            }
          }
        }
      }
    ],
    layout: "StandaloneLayout"
  });

  //</editor-fold>
};
