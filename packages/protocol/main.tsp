import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi";

import "./src/common/errors.tsp";
import "./src/auth/auth.tsp";
import "./src/session/session.tsp";

using TypeSpec.Http;
using TypeSpec.Rest;

@service({
  title: "MindHit API",
})
@server("http://localhost:8080", "Development server")
namespace MindHit;
